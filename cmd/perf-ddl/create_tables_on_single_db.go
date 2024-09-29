package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"try-out/pkg/config"
	"try-out/pkg/constants"
	"try-out/pkg/tidb"
)

func prepareForTableTest(db *sql.DB, dbName string) {
	_, err := db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	if err != nil {
		fmt.Printf("Failed to drop database %s: %v\n", dbName, err)
	} else {
		fmt.Printf("Dropped database %s\n", dbName)
	}
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	if err != nil {
		fmt.Printf("Failed to create database %s: %v\n", dbName, err)
	} else {
		fmt.Printf("Created database %s\n", dbName)
	}
}

func createTablesOnSingleDBAction(db *sql.DB) result {
	if config.GlobalCfg.Tables%config.GlobalCfg.Threads != 0 {
		panic(fmt.Sprintf("tables(%d) should be a multiple of threads(%d)", config.GlobalCfg.Tables, config.GlobalCfg.Threads))
	}
	if config.GlobalCfg.Databases != 1 {
		panic("databases must be 1")
	}

	if !config.GlobalCfg.SkipPrepare {
		prepareForTableTest(db, config.GlobalCfg.DatabaseName)
	}

	// wait log flush
	time.Sleep(2 * time.Second)
	executeStartTime := time.Now()
	fmt.Println("execute start time: ", executeStartTime.Format(logTimeFormat))

	totalTableCnt := config.GlobalCfg.Tables
	tablesPerThread := config.GlobalCfg.Tables / config.GlobalCfg.Threads
	var mu sync.Mutex
	durations := make([]time.Duration, 0, totalTableCnt)

	dbconns, err := tidb.PrepareConnections(db, config.GlobalCfg.Threads)
	if err != nil {
		fmt.Printf("Failed to prepare connections: %v\n", err)
		panic(err)
	}

	start := time.Now()
	var wg sync.WaitGroup
	for j := 0; j < config.GlobalCfg.Threads; j++ {
		idx := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := createTable(dbconns[idx], config.GlobalCfg.DatabaseName, idx, tablesPerThread)
			mu.Lock()
			durations = append(durations, res...)
			mu.Unlock()
		}()
	}
	wg.Wait()
	wallTime := time.Since(start)

	tidb.RecycleConnections(dbconns)

	fmt.Printf("\ntotal created %d tables, walltime: %v(%v per table)\n",
		totalTableCnt, wallTime.Round(time.Millisecond),
		(wallTime / time.Duration(totalTableCnt)).Round(time.Millisecond))
	printPercentile("action", durations)

	return result{
		startTime: executeStartTime,
	}
}

func createTable(db *sql.Conn, dbName string, idx int, tableCnt int) []time.Duration {
	durations := make([]time.Duration, tableCnt)
	for i := 0; i < tableCnt; i++ {
		tableName := fmt.Sprintf("tb_%d_%d", idx, i)
		tableCreateSQL := fmt.Sprintf(constants.DefaultCreateTableTemplate, dbName, tableName)
		st := time.Now()
		_, err := db.ExecContext(context.Background(), tableCreateSQL)
		if err != nil {
			fmt.Printf("Error creating table %s: %s\n", tableName, err.Error())
			panic(err)
		}
		durations[i] = time.Since(st)
	}
	return durations
}
