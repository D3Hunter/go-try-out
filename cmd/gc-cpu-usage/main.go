package main

import (
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/pprof"
	"regexp"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
	"unsafe"

	"github.com/docker/go-units"
	"github.com/pingcap/tidb/pkg/util/promutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"try-out/pkg/config"
)

type node struct {
	links [16]*node
}

func main() {
	nSize := int(unsafe.Sizeof(node{}))
	fmt.Println("node size:", nSize)
	go func() {
		if err := startServer(); err != nil && err != http.ErrServerClosed {
			config.GlobalLog.Error("start HTTP server failed", zap.Error(err))
		}
	}()
	// in total, we have 'CPU * 8 << 20 * 16' links
	// on a 16c machine, it's 2<<30 links
	concurrency := runtime.GOMAXPROCS(0)
	prevMemLimit := debug.SetMemoryLimit(25 * units.GiB)
	fmt.Println("previous memory limit:", units.HumanSize(float64(prevMemLimit)))
	fmt.Println("concurrency:", concurrency)
	fmt.Println("memory usage:", units.HumanSize(float64(nSize*4<<20*concurrency)))
	datas := make([]map[int]*node, concurrency)
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		i := i
		go func() {
			defer wg.Done()
			// 1GiB / 128 = 8 << 20
			count := 8 << 20
			m := make(map[int]*node, count)
			for j := 0; j < count; j++ {
				m[j] = &node{}
			}
			rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
			for j := 0; j < count; j++ {
				n := m[j]
				for k := 0; k < 16; k++ {
					n.links[k] = m[rnd.Intn(count)]
				}
			}
			datas[i] = m

			for i := 0; i < 1000; i++ {
				m[rnd.Intn(count)].links[rnd.Intn(16)] = m[rnd.Intn(count)]
				time.Sleep(time.Second)
			}
		}()
	}
	wg.Wait()
}

func startServer() error {
	mux := http.NewServeMux()
	registry := promutil.NewDefaultRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector(collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")})))
	if gatherer, ok := registry.(prometheus.Gatherer); ok {
		handler := promhttp.InstrumentMetricHandler(
			registry, promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{}),
		)
		mux.Handle("/metrics", handler)
	}

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	listener, err := net.Listen("tcp", ":10080")
	if err != nil {
		return err
	}
	config.GlobalLog.Info("starting HTTP server", zap.Stringer("address", listener.Addr()))
	var server http.Server
	server.Handler = mux
	return server.Serve(listener)
}
