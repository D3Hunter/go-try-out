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

func testInitMultipleTenantsAction(db *sql.DB) result {
	if config.GlobalCfg.Databases%config.GlobalCfg.Threads != 0 {
		panic(fmt.Sprintf("databases(%d) should be a multiple of threads(%d)", config.GlobalCfg.Databases, config.GlobalCfg.Threads))
	}
	if config.GlobalCfg.Tables%config.GlobalCfg.Databases != 0 {
		panic(fmt.Sprintf("tables（%d) should be a multiple of databases(%d)", config.GlobalCfg.Tables, config.GlobalCfg.Databases))
	}

	// wait log flush
	time.Sleep(2 * time.Second)
	executeStartTime := time.Now()
	fmt.Println("execute start time: ", executeStartTime.Format(logTimeFormat))

	tablesPerDB := config.GlobalCfg.Tables / config.GlobalCfg.Databases
	dbsPerThread := config.GlobalCfg.Databases / config.GlobalCfg.Threads
	var mu sync.Mutex
	createDBDurations := make([]time.Duration, 0, config.GlobalCfg.Databases)
	createTableDurations := make([]time.Duration, 0, config.GlobalCfg.Tables)

	dbconns, err := tidb.PrepareConnections(db, config.GlobalCfg.Threads)
	if err != nil {
		fmt.Printf("Failed to prepare connections: %v\n", err)
		panic(err)
	}

	start := time.Now()
	var wg sync.WaitGroup
	for j := 0; j < config.GlobalCfg.Threads; j++ {
		idx := j
		threadDBPrefix := fmt.Sprintf("%s_%d", config.GlobalCfg.DBPrefix, idx)
		wg.Add(1)
		go func() {
			defer wg.Done()
			dbDurations, tablesDurations := createTenant(dbconns[idx], threadDBPrefix, dbsPerThread, tablesPerDB)
			mu.Lock()
			createDBDurations = append(createDBDurations, dbDurations...)
			createTableDurations = append(createTableDurations, tablesDurations...)
			mu.Unlock()
		}()
	}
	wg.Wait()
	wallTime := time.Since(start)

	tidb.RecycleConnections(dbconns)

	fmt.Printf("\ntotal created %d databases, %d tables, walltime: %v(%v per tenant)\n",
		config.GlobalCfg.Databases, config.GlobalCfg.Tables, wallTime.Round(time.Millisecond),
		(wallTime / time.Duration(config.GlobalCfg.Databases)).Round(time.Millisecond))
	printPercentile("create-database", createDBDurations)
	printPercentile("create-table", createTableDurations)

	return result{
		startTime: executeStartTime,
	}
}

func createTenant(conn *sql.Conn, dbPrefix string, dbs int, tablesPerDB int) ([]time.Duration, []time.Duration) {
	createDBDurations := make([]time.Duration, 0, dbs)
	createTableDurations := make([]time.Duration, 0, dbs*tablesPerDB)
	ctx := context.Background()
	for i := 0; i < dbs; i++ {
		dbName := fmt.Sprintf("%s_%d", dbPrefix, i)
		st := time.Now()
		_, err := conn.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName))
		if err != nil {
			panic(err)
		}
		createDBDurations = append(createDBDurations, time.Since(st))

		for j := 0; j < tablesPerDB; j++ {
			tableName := fmt.Sprintf("tbl_%d", j)
			st = time.Now()
			tableCreateSQL := fmt.Sprintf(TableSQL, dbName, tableName)
			_, err = conn.ExecContext(ctx, tableCreateSQL)
			if err != nil {
				panic(err)
			}
			createTableDurations = append(createTableDurations, time.Since(st))
		}
	}
	return createDBDurations, createTableDurations
}
