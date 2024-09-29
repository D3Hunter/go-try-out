package test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/pingcap/tidb/pkg/kv"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

func TestJsonMarshal(t *testing.T) {
	bytes, err := json.Marshal(make(map[string]int))
	require.NoError(t, err)
	require.Equal(t, "{}", string(bytes))
	bytes, err = json.Marshal((map[string]int)(nil))
	require.NoError(t, err)
	require.Equal(t, "null", string(bytes))
}

func TestSetVarOnCompile(t *testing.T) {
	t.Logf(TestVar)
}

type testAtomicStruct struct {
	A atomic.Int64
	B atomic.Int64
}

type testNormalStruct struct {
	A int64
	B int64
}

func TestMarshalUberAtomic(t *testing.T) {
	t.Run("marshal atomic", func(t *testing.T) {
		ns := testNormalStruct{A: 123, B: 456}
		bytes, err := json.Marshal(ns)
		require.NoError(t, err)
		fmt.Println(string(bytes))
		s := testAtomicStruct{}
		err = json.Unmarshal(bytes, &s)
		require.NoError(t, err)
		fmt.Println(s.A.Load(), s.B.Load())
		bytes, err = json.Marshal(&s)
		require.NoError(t, err)
		fmt.Println(string(bytes))
	})
	t.Run("marshal 2", func(t *testing.T) {
		s := testAtomicStruct{}
		s.A.Store(123)
		s.B.Store(456)
		bytes, err := json.Marshal(&s)
		require.NoError(t, err)
		fmt.Println(string(bytes))
	})
}

func TestSelectClosedCh(t *testing.T) {
	ch := make(chan struct{})
	close(ch)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second)
		cancel()
	}()
	for i := 0; i < 50; i++ {
		//select {
		//case <-ctx.Done():
		//	fmt.Println("from context done")
		//case <-ch:
		//	fmt.Println("from closed ch")
		//	time.Sleep(300 * time.Millisecond)
		//}

		select {
		case <-ctx.Done():
			fmt.Println("from context done")
		default:
			fmt.Println("from default")
			time.Sleep(300 * time.Millisecond)
		}
	}
}

func TestDiff(t *testing.T) {
	const (
		text1 = "Lorem ipsum dolor."
		text2 = "Lorem dolor sit amet."
	)

	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(text1, text2, false)

	fmt.Println(dmp.DiffPrettyText(diffs))
	fmt.Println(diffs)
}

func TestPrintX(t *testing.T) {
	data := []byte{0x74, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x6e, 0x5f, 0x72, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04}
	fmt.Printf("%x\n", data)
	k := kv.Key(data)
	fmt.Printf("%x\n", k)
	fmt.Printf("%x\n", 9223090561878065153)
}
