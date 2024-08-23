package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"
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

func TestIf(tt *testing.T) {
	x := &a{}
	x.F()
	replace((&x).(*t))
	x.F()
}
