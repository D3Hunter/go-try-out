package test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func TestEtcdWatch(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://localhost:2379"},
	})
	require.NoError(t, err)
	watchChan := client.Watch(context.Background(), "xxxxxxx")
	t.Log(<-watchChan)
}

func TestEtcdWatchPutWithoutChange(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://localhost:2379"},
	})
	require.NoError(t, err)
	watchChan := client.Watch(context.Background(), "xxxxxxx")
	for resp := range watchChan {
		bytes, err := json.Marshal(resp)
		require.NoError(t, err)
		t.Log(string(bytes))
	}
}

func TestEtcdLeaseDeleted(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://localhost:2379"},
	})
	require.NoError(t, err)
	session, err := concurrency.NewSession(client, concurrency.WithTTL(100))
	require.NoError(t, err)
	t.Logf("session id: %x", session.Lease())
	<-session.Done()
	t.Log("session done")
	require.NoError(t, session.Close())
}

func TestWatchResponse(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://localhost:2379"},
	})
	require.NoError(t, err)
	ctx := context.Background()
	resp, err := client.Put(ctx, "/key", "a")
	require.NoError(t, err)
	_, err = client.Put(ctx, "/key", "a")
	require.NoError(t, err)
	_, err = client.Put(ctx, "/key", "a")
	require.NoError(t, err)
	watch := client.Watch(ctx, "/key", clientv3.WithRev(resp.Header.Revision))
	wResp := <-watch
	for _, e := range wResp.Events {
		fmt.Println(e.Type)
	}
}

func TestEtcdDeleteAndPut(t *testing.T) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"http://localhost:2379"},
	})
	require.NoError(t, err)
	ctx := context.Background()
	prefix := "/ut/etcd/txn/"
	_, err = client.Delete(ctx, prefix, clientv3.WithPrefix())
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		_, err = client.Put(ctx, fmt.Sprintf("%s%d", prefix, i), "a")
		require.NoError(t, err)
	}
	getResp, err := client.Get(ctx, prefix, clientv3.WithPrefix())
	require.NoError(t, err)
	require.Len(t, getResp.Kvs, 3)
	newKey := prefix + "a"
	_, err = client.Txn(ctx).
		Then(
			clientv3.OpDelete(prefix, clientv3.WithPrefix()),
			clientv3.OpPut(newKey, "a"),
		).Commit()
	require.NoError(t, err)
	getResp, err = client.Get(ctx, prefix+"/", clientv3.WithPrefix())
	require.NoError(t, err)
	require.Len(t, getResp.Kvs, 1)
	require.EqualValues(t, newKey, string(getResp.Kvs[0].Key))
}

type t interface {
	F()
}

type a struct{}

func (*a) F() { fmt.Println("a") }

type b struct{}

func (*b) F() { fmt.Println("b") }

func replace(tp *t) {
	*tp = &b{}
}
