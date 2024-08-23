package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

func createDatabasesAction(db *sql.DB) result {
	if globalCfg.databases%globalCfg.threads != 0 {
		panic(fmt.Sprintf("databases(%d) should be a multiple of threads(%d)", globalCfg.databases, globalCfg.threads))
	}
	// wait log flushed
	time.Sleep(2 * time.Second)
	executeStartTime := time.Now()
	fmt.Println("execute start time: ", executeStartTime.Format(logTimeFormat))

	createFn := func(conn *sql.Conn, idx int, cnt int) []time.Duration {
		durations := make([]time.Duration, cnt)
		for i := 0; i < cnt; i++ {
			sql := fmt.Sprintf("create database %s_%d_%d", globalCfg.dbPrefix, idx, i)
			st := time.Now()
			_, err := conn.ExecContext(context.Background(), sql)
			if err != nil {
				fmt.Printf("Error on %s: %s\n", sql, err.Error())
			}
			durations[i] = time.Since(st)
		}
		return durations
	}
	conns, err := prepareConnections(db, globalCfg.threads)
	defer recycleConnections(conns)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	start := time.Now()
	dbPerThread := globalCfg.databases / globalCfg.threads
	durations := make([]time.Duration, 0, globalCfg.databases)
	for i := 0; i < globalCfg.threads; i++ {
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
		globalCfg.databases, wallTime.Round(time.Millisecond),
		(wallTime / time.Duration(globalCfg.databases)).Round(time.Millisecond))
	printPercentile("action", durations)

	return result{
		startTime: executeStartTime,
	}
}
