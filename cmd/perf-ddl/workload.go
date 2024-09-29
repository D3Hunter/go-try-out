package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"try-out/pkg/config"
	"try-out/pkg/tidb"
)

func workloadAction(db *sql.DB) result {
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
		wg.Add(1)
		go func() {
			defer wg.Done()
			dbDurations, tablesDurations := querySomeTable(dbconns[idx], idx, tablesPerDB)
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

func querySomeTable(conn *sql.Conn, id, tablesPerDB int) ([]time.Duration, []time.Duration) {
	ctx := context.Background()
	rnd := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
	for {
		count := 0
		dbName := fmt.Sprintf("%s_%d", config.GlobalCfg.DBPrefix, rnd.Intn(config.GlobalCfg.Databases))
		for j := 0; j < tablesPerDB; j++ {
			tableName := fmt.Sprintf("tbl_%d", j)
			sel := fmt.Sprintf("select count(1) from %s.%s", dbName, tableName)
			rows, err := conn.QueryContext(ctx, sel)
			if err != nil {
				if strings.Contains(err.Error(), "doesn't exist") {
					break
				}
				panic(err)
			} else {
				count++
				rows.Close()
				_, err := conn.ExecContext(ctx, fmt.Sprintf("insert into %s.%s(column_1,column_13,column_14,column_15,column_16) values(%d,now(),now(),now(),now())", dbName, tableName, rnd.Intn(1<<20)))
				if err != nil {
					panic(err)
				}
			}
		}
		fmt.Println("tables iterated:", id, count)
	}
	return nil, nil
}
