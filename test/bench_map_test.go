package test

import (
	"testing"

	"github.com/pingcap/parser/model"
)

type testStruct struct {
	ai *model.TableInfo
	bi *model.DBInfo
}

func BenchmarkMap1(b *testing.B) {
	ti := &model.TableInfo{
		ID: 1,
	}
	di := &model.DBInfo{
		ID: 1,
	}
	count := 1024
	b.Run("slice iter", func(b *testing.B) {
		sum := int64(0)
		var (
			ts *testStruct
		)
		for i := 0; i < b.N; i++ {
			s := make([]*testStruct, 0, count)
			for k := 0; k < count; k++ {
				ts = new(testStruct)
				s = append(s, ts)
				for i := 0; i < len(s); i++ {
					if s[i] == ts {
						sum++
					}
				}
			}
		}
		_ = sum
		_ = ts
	})
	b.Run("open addressing map", func(b *testing.B) {
		sum := int64(0)
		var (
			ts *testStruct
		)
		for i := 0; i < b.N; i++ {
			m := New[*testStruct, int64](count)
			for k := 0; k < count; k++ {
				ts = new(testStruct)
				m.Set(ts, 1)
				v, _ := m.Get(ts)
				sum += v
			}
		}
		_ = sum
		_ = ts
	})
	b.Run("go map size", func(b *testing.B) {
		sum := int64(0)
		var (
			ts *testStruct
			m  map[any]int64
		)
		for i := 0; i < b.N; i++ {
			m = make(map[any]int64, count)
			for k := 0; k < count; k++ {
				ts = new(testStruct)
				m[ts] = 1
				sum += m[ts]
			}
		}
		_ = sum
		_, _ = ts, m
	})
	b.Run("go map def", func(b *testing.B) {
		sum := int64(0)
		var (
			ts *testStruct
			m  map[any]int64
		)
		for i := 0; i < b.N; i++ {
			m = make(map[any]int64)
			for k := 0; k < count; k++ {
				ts = new(testStruct)
				m[ts] = 1
				sum += m[ts]
			}
		}
		_ = sum
		_, _ = ts, m
	})
	b.Run("normal", func(b *testing.B) {
		sum := int64(0)
		var lastP *model.TableInfo
		var (
			ts *testStruct
		)
		for i := 0; i < b.N; i++ {
			for k := 0; k < count; k++ {
				ts = new(testStruct)
				ts.ai = ti
				ts.bi = di
				sum += ts.ai.ID
				sum += ts.bi.ID
				lastP = ts.ai
			}
		}
		_ = lastP
	})
}

func BenchmarkMapRead(b *testing.B) {
	count := 32
	lastTs := new(testStruct)
	s := make([]*testStruct, 0, count)
	m := make(map[any]int64, count)
	for k := 0; k < count; k++ {
		ts := new(testStruct)
		m[ts] = 1
		s = append(s, ts)
		lastTs = ts
	}
	b.Run("slice iter", func(b *testing.B) {
		sum := int64(0)
		for i := 0; i < b.N; i++ {
			for k := 0; k < count; k++ {
				for i := 0; i < len(s); i++ {
					if s[i] == lastTs {
						sum++
					}
				}
			}
		}
		_ = sum
	})
	b.Run("go map size", func(b *testing.B) {
		sum := int64(0)
		for i := 0; i < b.N; i++ {
			for k := 0; k < count; k++ {
				sum += m[lastTs]
			}
		}
		_ = sum
	})
}
