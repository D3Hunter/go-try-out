package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	tidbkv "github.com/pingcap/tidb/pkg/kv"
	"github.com/stretchr/testify/require"
	"github.com/tikv/client-go/v2/config"
	kverror "github.com/tikv/client-go/v2/error"
	"github.com/tikv/client-go/v2/kv"
	"github.com/tikv/client-go/v2/oracle"
	"github.com/tikv/client-go/v2/txnkv"
	"go.uber.org/atomic"
)

// Init initializes information.
func initStore(t *testing.T) *txnkv.Client {
	client, err := txnkv.NewClient([]string{"127.0.0.1:2379"})
	require.NoError(t, err)
	return client
}

func beginTxn(t *testing.T, client *txnkv.Client, pessimistic bool) (txn *txnkv.KVTxn) {
	txn, err := client.Begin()
	require.NoError(t, err)
	txn.SetPessimistic(pessimistic)
	return txn
}

func TestExampleForPessimisticTXN(t *testing.T) {
	ctx := context.Background()
	client := initStore(t)

	key := []byte("hello")
	readTxn := beginTxn(t, client, false)
	_, err := readTxn.Get(ctx, key)
	require.ErrorIs(t, err, kverror.ErrNotExist)
	require.NoError(t, readTxn.Commit(ctx))

	t.Cleanup(func() {
		delTxn := beginTxn(t, client, false)
		require.NoError(t, delTxn.Delete(key))
		require.NoError(t, delTxn.Commit(ctx))
	})

	getVal := func(k []byte) []byte {
		txn := beginTxn(t, client, false)
		val, err := txn.Get(ctx, k)
		require.NoError(t, err)
		return val
	}

	t.Run("pessimistic-lock", func(t *testing.T) {
		txn1 := beginTxn(t, client, true)
		t.Log("before lock", time.Now())
		require.NoError(t, txn1.LockKeysWithWaitTime(ctx, kv.LockAlwaysWait, key))
		t.Log("before sleep", time.Now())
		t.Log(os.Args)
		if len(os.Args) > 4 && os.Args[4] == "sleep" {
			time.Sleep(10 * time.Second)
		}
		t.Log("after sleep", time.Now())

		txn2 := beginTxn(t, client, true)
		err = txn2.LockKeysWithWaitTime(ctx, kv.LockNoWait, key)
		require.ErrorIs(t, err, kverror.ErrLockAcquireFailAndNoWaitSet)
		err = txn2.LockKeysWithWaitTime(ctx, 200, key)
		require.ErrorIs(t, err, kverror.ErrLockWaitTimeout)

		// commit txn1 should be success.
		require.NoError(t, txn1.Set(key, []byte("world")))
		require.NoError(t, txn1.Commit(ctx))
		require.EqualValues(t, "world", getVal(key))

		require.NoError(t, txn2.LockKeysWithWaitTime(ctx, kv.LockNoWait, key))
		require.NoError(t, txn2.Set(key, []byte("new world")))
		require.NoError(t, txn2.Commit(ctx))
		require.EqualValues(t, "new world", getVal(key))
	})

	t.Run("pessimistic-lock-expire", func(t *testing.T) {
		config.GetGlobalConfig().MaxTxnTTL = 5000
		txn1 := beginTxn(t, client, true)
		require.NoError(t, txn1.LockKeysWithWaitTime(ctx, 5000, key))

		time.Sleep(7 * time.Second)
		txn2 := beginTxn(t, client, true)
		err = txn2.LockKeysWithWaitTime(ctx, kv.LockAlwaysWait, key)
		require.NoError(t, err)

		// commit txn1 should be success.
		//require.NoError(t, txn1.Set(key, []byte("world")))
		//require.NoError(t, txn1.Commit(ctx))
		//require.EqualValues(t, "world", getVal(key))

		//require.NoError(t, txn2.LockKeysWithWaitTime(ctx, kv.LockNoWait, key))
		require.NoError(t, txn2.Set(key, []byte("new world")))
		require.NoError(t, txn2.Commit(ctx))
		require.EqualValues(t, "new world", getVal(key))
	})

	t.Run("pessimistic-no-lock", func(t *testing.T) {
		txn1 := beginTxn(t, client, true)
		require.NoError(t, txn1.Set(key, []byte("world")))

		txn2 := beginTxn(t, client, true)
		require.NoError(t, txn2.Set(key, []byte("new world")))

		require.NoError(t, txn1.Commit(ctx))
		require.EqualValues(t, "world", getVal(key))

		require.NoError(t, txn2.Commit(ctx))
		require.EqualValues(t, "new world", getVal(key))
	})
}

func lockKeyWithRetry(ctx context.Context, t *testing.T, client *txnkv.Client, txn *txnkv.KVTxn, key []byte) (forUpdateTs uint64) {
	for i := 0; i < 100; i++ {
		var err error
		forUpdateTs = txn.StartTS()
		if txn.IsPessimistic() {
			forUpdateTs, err = client.CurrentTimestamp(oracle.GlobalTxnScope)
			if err != nil {
				panic(err)
			}
		}
		lockCtx := kv.NewLockCtx(forUpdateTs, 30000, time.Now())
		if err = txn.LockKeys(ctx, lockCtx, key); err != nil {
			if kverror.IsErrWriteConflict(err) {
				tidbkv.BackOff(uint(i))
				continue
			}
			panic(err)
		}
		break
	}
	return forUpdateTs
}

func TestQPSPessimisticTxn(t *testing.T) {
	ctx := context.Background()
	client := initStore(t)

	dataVal := make([]byte, 5024)
	for i := 0; i < len(dataVal); i++ {
		dataVal[i] = 'a'
	}
	idKey := []byte("INCR_ID_KEY")
	var wg sync.WaitGroup
	thread := 1000
	iterationPerThread := 30000
	counters := make([]atomic.Int64, thread+1)
	for i := 0; i < thread; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ids := make([]int64, iterationPerThread)
			for j := 0; j < iterationPerThread; j++ {
				txn := beginTxn(t, client, true)
				forUpdateTS := lockKeyWithRetry(ctx, t, client, txn, idKey)
				txn.GetSnapshot().SetSnapshotTS(forUpdateTS)
				val, err := txn.Get(ctx, idKey)
				if err == kverror.ErrNotExist {
					val = []byte("0")
				} else if err != nil {
					panic(err)
				}
				intVal, err := strconv.ParseInt(string(val), 10, 64)
				if err != nil {
					panic(err)
				}
				if err = txn.Set(idKey, []byte(strconv.FormatInt(intVal+1, 10))); err != nil {
					panic(err)
				}
				//dataKey := fmt.Sprintf("data_key_%d", intVal)
				//if err = txn.Set([]byte(dataKey), dataVal); err != nil {
				//	panic(err)
				//}
				if err = txn.Commit(ctx); err != nil {
					panic(err)
				}
				counters[0].Add(1)
				counters[index+1].Add(1)
				ids[j] = intVal
			}
			fmt.Printf("thread-%d ids: %v\n", index, ids[:100])
		}(i)
	}
	go func() {
		getCounts := func() []int64 {
			res := make([]int64, len(counters))
			for i := range counters {
				res[i] = counters[i].Load()
			}
			return res
		}
		lastCnt := getCounts()
		for {
			select {
			case <-time.After(5 * time.Second):
			}
			currCnt := getCounts()
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("QPS - total:%.0f", float64(currCnt[0]-lastCnt[0])/5))
			for i := 1; i < len(counters); i++ {
				sb.WriteString(fmt.Sprintf(", thread-%d: %.0f", i, float64(currCnt[i]-lastCnt[i])/5))
			}
			lastCnt = currCnt
			fmt.Println(sb.String())
		}
	}()
	wg.Wait()
	readTxn := beginTxn(t, client, false)
	val, err := readTxn.Get(ctx, idKey)
	require.NoError(t, err)
	require.EqualValues(t, strconv.Itoa(iterationPerThread*thread), string(val))
	require.NoError(t, readTxn.Commit(ctx))
}
