package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

func showVersion(db *sql.DB) error {
	rs, err := db.Query("select tidb_version()")
	if err != nil {
		return err
	}
	defer rs.Close()
	if rs.Next() {
		var verStr string
		if err := rs.Scan(&verStr); err != nil {
			return err
		}
		fmt.Printf("TiDB version: %s\n", verStr)
	}
	return rs.Err()
}

func getAllDatabases(db *sql.DB) ([]string, error) {
	rs, err := db.Query("show databases")
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	res := make([]string, 0, 100)
	for rs.Next() {
		var dbName string
		if err = rs.Scan(&dbName); err != nil {
			return nil, err
		}
		res = append(res, dbName)
	}
	return res, rs.Err()
}

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
	dbs, err := getAllDatabases(db)
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

func prepareConnections(db *sql.DB, thread int) ([]*sql.Conn, error) {
	dbconns := make([]*sql.Conn, 0, thread)

	for i := 0; i < thread; i++ {
		conn, err := db.Conn(context.Background())
		if err != nil {
			return nil, err
		}
		dbconns = append(dbconns, conn)
	}
	return dbconns, nil
}

func recycleConnections(conns []*sql.Conn) {
	for _, conn := range conns {
		conn.Close()
	}
}

func truncateTable(db *sql.Conn, dbName string, idx int, tableCnt int) []time.Duration {
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
