package main

import (
	"cmp"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"

	"github.com/pingcap/tidb/pkg/kv"
	"github.com/pingcap/tidb/pkg/tablecodec"
	"github.com/pingcap/tidb/pkg/util/codec"
)

type Regions struct {
	Count   int `json:"count"`
	Regions []Region
}

type Region struct {
	ID       uint64 `json:"id"`
	StartKey string `json:"start_key"`
	EndKey   string `json:"end_key"`
	Peers    []Peer `json:"peers"`
}

type Peer struct {
	ID      uint64 `json:"id"`
	StoreID uint64 `json:"store_id"`
}

type storeRegionCnt struct {
	storeID uint64
	cnt     int
}

func (s storeRegionCnt) String() string {
	return fmt.Sprintf("%2d:%2d", s.storeID, s.cnt)
}

func main() {
	//
	// change this depends on your range-concurrency, suppose it's 25, set it to 50
	//
	var step = 200
	//
	// target table id
	//
	tableID := int64(128)
	//
	// you can get it from PD using: curl localhost:2379/pd/api/v1/regions
	// ******* Note: this file might be quite large, as it include all regions in the cluster. ******
	//
	content, err := os.ReadFile("/path/to/regions.json")
	if err != nil {
		panic(err)
	}
	prefixBytes := codec.EncodeBytes(nil, tablecodec.EncodeRowKeyWithHandle(tableID, kv.IntHandle(0)))[:10]
	prefix := hex.EncodeToString(prefixBytes)
	if err != nil {
		panic(err)
	}
	regions := Regions{}
	if err := json.Unmarshal(content, &regions); err != nil {
		panic(err)
	}
	tableRegions := make([]Region, 0, len(regions.Regions))
	for _, r := range regions.Regions {
		if strings.HasPrefix(r.StartKey, prefix) {
			tableRegions = append(tableRegions, r)
		}
	}
	slices.SortFunc(tableRegions, func(a, b Region) int {
		return strings.Compare(a.StartKey, b.StartKey)
	})

	for i := 0; i < len(tableRegions); i += step {
		end := min(i+step, len(tableRegions))
		distributions := make(map[uint64]int)
		for j := i; j < end; j++ {
			for _, p := range tableRegions[j].Peers {
				distributions[p.StoreID]++
			}
		}
		storeRegionCnts := make([]storeRegionCnt, 0, len(distributions))
		for storeID, cnt := range distributions {
			storeRegionCnts = append(storeRegionCnts, storeRegionCnt{storeID, cnt})
		}
		slices.SortFunc(storeRegionCnts, func(a, b storeRegionCnt) int {
			return cmp.Compare(a.storeID, b.storeID)
		})
		//fmt.Printf("distributions for [%s, %s): %v\n",
		//	tableRegions[i].StartKey, tableRegions[end-1].EndKey,
		//	storeRegionCnts)
		var sum int
		for _, rc := range storeRegionCnts {
			sum += rc.cnt
		}
		avg := float64(sum) / float64(len(storeRegionCnts))
		var varSum float64
		for _, rc := range storeRegionCnts {
			varSum += math.Pow(float64(rc.cnt)-avg, 2)
		}
		sdv := math.Sqrt(varSum / float64(len(storeRegionCnts)))
		fmt.Printf("distributions: avg:%.2f, sdv %.2f, %v\n",
			avg, sdv, storeRegionCnts)
	}
}
