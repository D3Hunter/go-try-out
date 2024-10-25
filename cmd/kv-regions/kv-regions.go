package main

import (
	"fmt"

	"github.com/pingcap/tidb/pkg/kv"
	"github.com/pingcap/tidb/pkg/tablecodec"
	"github.com/pingcap/tidb/pkg/util/codec"
	"try-out/pkg/constants"
)

func main() {
	_ = codec.EncodeBytes(nil, tablecodec.EncodeRowKeyWithHandle(1, kv.IntHandle(0)))[:10]
	fmt.Println(constants.GlobalVar)
}
