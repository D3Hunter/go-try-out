package test

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pingcap/tidb/pkg/meta"
	"github.com/pingcap/tidb/pkg/structure"
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

func TestEncodeKey(t *testing.T) {
	for _, s := range []string{
		"333334303032",
		"3337393239303438",
	} {
		res, err := hex.DecodeString(s)
		require.NoError(t, err)
		fmt.Println(string(res))
	}
	txn := structure.NewStructure(nil, nil, []byte("m"))
	key := txn.EncodeHashDataKey(meta.DBkey(167), meta.AutoIncrementIDKey(169))
	fmt.Println(hex.EncodeToString(key))
}
