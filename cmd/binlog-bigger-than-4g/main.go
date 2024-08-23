package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type testif interface {
	test()
}

type impl int

func (i impl) test() {
	fmt.Printf("current val is %d", i)
	i = 2
	fmt.Println("set to 2")
}

func test_mysql() {
	//db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/test?loc=Local")
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:4000)/test")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		panic(err)
	}
	rows, err := tx.Query("select count(*) from t where name='test'")
	if err != nil {
		panic(err)
	}
	var count int
	rows.Next()
	err = rows.Scan(&count)
	if err != nil {
		panic(err)
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}
	_ = rows.Close()
	if count > 0 {
		_ = tx.Rollback()
		return
	}
	_, err = tx.Exec("insert into t values('test')")
	if err != nil {
		panic(err)
	}
	err = tx.Commit()
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println(time.Since(time.Date(2022, 11, 9, 9, 14, 0, 0, time.UTC)))
}
