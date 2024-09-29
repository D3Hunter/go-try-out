package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"try-out/pkg/config"
	"try-out/pkg/tidb"
)

func createDatabasesAction(db *sql.DB) result {
	if config.GlobalCfg.Databases%config.GlobalCfg.Threads != 0 {
		panic(fmt.Sprintf("databases(%d) should be a multiple of threads(%d)", config.GlobalCfg.Databases, config.GlobalCfg.Threads))
	}
	// wait log flushed
	time.Sleep(2 * time.Second)
	executeStartTime := time.Now()
	fmt.Println("execute start time: ", executeStartTime.Format(logTimeFormat))

	createFn := func(conn *sql.Conn, idx int, cnt int) []time.Duration {
		durations := make([]time.Duration, cnt)
		for i := 0; i < cnt; i++ {
			sql := fmt.Sprintf("create database %s_%d_%d", config.GlobalCfg.DBPrefix, idx, i)
			st := time.Now()
			_, err := conn.ExecContext(context.Background(), sql)
			if err != nil {
				fmt.Printf("Error on %s: %s\n", sql, err.Error())
			}
			durations[i] = time.Since(st)
		}
		return durations
	}
	conns, err := tidb.PrepareConnections(db, config.GlobalCfg.Threads)
	defer tidb.RecycleConnections(conns)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	start := time.Now()
	dbPerThread := config.GlobalCfg.Databases / config.GlobalCfg.Threads
	durations := make([]time.Duration, 0, config.GlobalCfg.Databases)
	for i := 0; i < config.GlobalCfg.Threads; i++ {
		idx := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := createFn(conns[idx], idx, dbPerThread)
			mu.Lock()
			durations = append(durations, res...)
			mu.Unlock()
		}()
	}
	wg.Wait()
	wallTime := time.Since(start)

	fmt.Printf("\ntotal created %d databases, wallTime: %v(%v per op)\n",
		config.GlobalCfg.Databases, wallTime.Round(time.Millisecond),
		(wallTime / time.Duration(config.GlobalCfg.Databases)).Round(time.Millisecond))
	printPercentile("action", durations)

	return result{
		startTime: executeStartTime,
	}
}
