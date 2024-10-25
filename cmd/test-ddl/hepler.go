package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"try-out/pkg/tidb"
)

var systemDatabases = map[string]struct{}{
	"MYSQL":              {},
	"INFORMATION_SCHEMA": {},
	"PERFORMANCE_SCHEMA": {},
	"METRICS_SCHEMA":     {},
	"SYS":                {},
}

func prepareForDatabaseTest(db *sql.DB) error {
	start := time.Now()
	defer func() {
		fmt.Printf("prepareForDatabaseTest takes %v\n", time.Since(start))
	}()
	dbs, err := tidb.GetAllDatabases(db)
	if err != nil {
		return err
	}

	eg, _ := errgroup.WithContext(context.Background())
	eg.SetLimit(16)
	for _, dbName := range dbs {
		if _, ok := systemDatabases[strings.ToUpper(dbName)]; ok {
			continue
		}
		dbName := dbName
		eg.Go(func() error {
			_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
			return err
		})
	}
	return eg.Wait()
}

func cleanUp(host string, port, databaseCnt int, dbPrefix string) {
	db, err := sql.Open("mysql", fmt.Sprintf("root@tcp(%s:%d)/", host, port))
	if err != nil {
		fmt.Printf("Failed to connect to MySQL database: %v\n", err)
		return
	}
	defer db.Close()
	for i := 0; i < databaseCnt; i++ {
		_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s_%d", dbPrefix, i))
		if err != nil {
			fmt.Printf("Failed to drop database %s_%d: %v\n", dbPrefix, i, err)
		} else {
			fmt.Printf("Dropped database %s_%d\n", dbPrefix, i)
		}
	}
}

func TruncateTable(db *sql.Conn, dbName string, idx int, tableCnt int) []time.Duration {
	durations := make([]time.Duration, tableCnt)
	for i := 0; i < tableCnt; i++ {
		tableName := fmt.Sprintf("tb_%d_%d", idx, i)
		tableCreateSQL := fmt.Sprintf("truncate table %s.%s", dbName, tableName)
		st := time.Now()
		_, err := db.ExecContext(context.Background(), tableCreateSQL)
		if err != nil {
			fmt.Printf("Error creating table %s: %s\n", tableName, err.Error())
		}
		durations[i] = time.Since(st)
	}
	return durations
}
