package test

import (
	"math/rand"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"testing"
	"time"

	"github.com/pingcap/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func InitStatus() {
	lis, err := net.Listen("tcp", "0.0.0.0:3333")
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	httpS := &http.Server{
		Handler: mux,
	}
	err = httpS.Serve(lis)
	if err != nil {
		log.L().Error("status server returned", zap.Error(err))
	}
}

func TestPrometheus(t *testing.T) {
	registry := prometheus.NewRegistry()
	gaugeVec := prometheus.NewGaugeVec(prometheus.GaugeOpts{Namespace: "tryout2", Name: "state"}, []string{"jobid", "state"})
	registry.MustRegister(gaugeVec)
	prometheus.DefaultGatherer = registry
	failedGauges, runningGauges := []prometheus.Gauge{}, []prometheus.Gauge{}
	for i := 0; i < 10; i++ {
		gauge := gaugeVec.WithLabelValues(strconv.Itoa(i), "failed")
		failedGauges = append(failedGauges, gauge)
		gauge = gaugeVec.WithLabelValues(strconv.Itoa(i), "running")
		runningGauges = append(runningGauges, gauge)
	}

	go func() {
		values := []int{0, 0, 0, 0, 1}
		successVals := []int{1, 1, 1, 0}
		for i := 0; i < 100; i++ {
			time.Sleep(2 * time.Second)
			for _, g := range failedGauges {
				g.Set(float64(values[rand.Int()%len(values)]))
			}
			for _, g := range runningGauges {
				g.Set(float64(successVals[rand.Int()%len(successVals)]))
			}
		}
	}()
	InitStatus()
}
