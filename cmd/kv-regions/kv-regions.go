package main

import (
	"github.com/pingcap/tidb/pkg/kv"
	"github.com/pingcap/tidb/pkg/tablecodec"
	"github.com/pingcap/tidb/pkg/util/codec"
)

func main() {
	_ = codec.EncodeBytes(nil, tablecodec.EncodeRowKeyWithHandle(1, kv.IntHandle(0)))[:10]
}
