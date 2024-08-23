package main

import (
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pingcap/tidb/pkg/util/codec"
	"github.com/stretchr/testify/require"
)

func TestTiDBKey2TiKVFormat(t *testing.T) {
	t.Log(time.Now().Add(-1))
	t.Logf((45 * time.Second).String())
	s := []byte("748000fffffffffffe5f72800000000000022f")
	dst := make([]byte, hex.DecodedLen(len(s)))
	_, err := hex.Decode(dst, s)
	require.NoError(t, err)
	strconv.Quote(string(dst))
	dst = codec.EncodeBytes(nil, dst)
	var res strings.Builder
	for _, c := range dst {
		if 0x20 <= c && c <= 0x7E {
			res.WriteByte(c)
			continue
		}
		res.WriteString(`\`)
		octal := strconv.FormatInt(int64(c), 8)
		for i := 0; i < 3-len(octal); i++ {
			res.WriteString("0")
		}
		res.WriteString(octal)
	}
	t.Logf(res.String())
}
