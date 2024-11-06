package tidb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func ShowVersion(db *sql.DB) error {
	rs, err := db.Query("select version()")
	if err != nil {
		return err
	}
	defer rs.Close()
	var verStr string
	if rs.Next() {
		if err = rs.Scan(&verStr); err != nil {
			return err
		}
		fmt.Printf("Version: %s\n", verStr)
	}
	if err = rs.Err(); err != nil {
		return err
	}
	if strings.Contains(verStr, "-TiDB-") {
		return showTiDBVersion(db)
	}
	return nil
}

func showTiDBVersion(db *sql.DB) error {
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

func GetAllDatabases(db *sql.DB) ([]string, error) {
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

func PrepareConnections(db *sql.DB, thread int) ([]*sql.Conn, error) {
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

func RecycleConnections(conns []*sql.Conn) {
	for _, conn := range conns {
		conn.Close()
	}
}
