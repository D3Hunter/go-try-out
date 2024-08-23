package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

func tableLevelDDLAction(db *sql.DB) result {
	if !strings.Contains(globalCfg.tableDDLTemplate, "%s.%s") {
		panic("tableDDLTemplate must contain '%s.%s'")
	}
	if globalCfg.tables%globalCfg.threads != 0 {
		panic(fmt.Sprintf("tables(%d) should be a multiple of threads(%d)", globalCfg.tables, globalCfg.threads))
	}
	if globalCfg.databases != 1 {
		panic("databases must be 1")
	}

	// wait log flush
	time.Sleep(2 * time.Second)
	executeStartTime := time.Now()
	fmt.Println("execute start time: ", executeStartTime.Format(logTimeFormat))

	totalTableCnt := globalCfg.tables
	tablesPerThread := globalCfg.tables / globalCfg.threads
	var mu sync.Mutex
	durations := make([]time.Duration, 0, totalTableCnt)

	dbconns, err := prepareConnections(db, globalCfg.threads)
	if err != nil {
		fmt.Printf("Failed to prepare connections: %v\n", err)
		panic(err)
	}

	start := time.Now()
	var wg sync.WaitGroup
	for j := 0; j < globalCfg.threads; j++ {
		idx := j
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := tableLevelDDL(dbconns[idx], globalCfg.databaseName, idx, tablesPerThread)
			mu.Lock()
			durations = append(durations, res...)
			mu.Unlock()
		}()
	}
	wg.Wait()
	wallTime := time.Since(start)

	recycleConnections(dbconns)

	fmt.Printf("\ntotal done %d DDLs, walltime: %v(%v per DDL)\n",
		totalTableCnt, wallTime.Round(time.Millisecond),
		(wallTime / time.Duration(totalTableCnt)).Round(time.Millisecond))
	printPercentile("action", durations)

	return result{
		startTime: executeStartTime,
	}
}

func tableLevelDDL(db *sql.Conn, dbName string, idx int, tableCnt int) []time.Duration {
	durations := make([]time.Duration, tableCnt)
	for i := 0; i < tableCnt; i++ {
		tableName := fmt.Sprintf("tb_%d_%d", idx, i)
		s := fmt.Sprintf(globalCfg.tableDDLTemplate, dbName, tableName)
		st := time.Now()
		_, err := db.ExecContext(context.Background(), s)
		if err != nil {
			fmt.Printf("Error on table level DDL %s: %s\n", s, err.Error())
			panic(err)
		}
		durations[i] = time.Since(st)
	}
	return durations
}
