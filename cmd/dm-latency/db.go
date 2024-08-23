package main

import (
	"context"
	"database/sql"
)

func OpenDbConn(url string) (*sql.Conn, error) {
	db, err := sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
