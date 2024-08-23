package main

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

type Generator struct {
	conn *sql.Conn
}

func (g *Generator) Init() error {
	var err error
	if g.conn, err = OpenDbConn("root:123456@tcp(127.0.0.1:3306)/test?loc=Local"); err != nil {
		return err
	}
	return nil
}

func (g *Generator) Start(wg *sync.WaitGroup) {
	g.StartInsert(wg)
}

func (g *Generator) StartInsert(wg *sync.WaitGroup) {
	defer wg.Done()
	//stmt, err := g.conn.PrepareContext(context.Background(), "update ts_test set ts=current_timestamp(6) where id=1")
	//stmt, err := g.conn.PrepareContext(context.Background(), "update ts_test set ts=? where id=1")
	stmt, err := g.conn.PrepareContext(context.Background(), "insert into ts_test(ts, ts2) values(?, ?)")
	if err != nil {
		panic(err)
	}
	for {
		//if _, err = stmt.Exec(); err != nil {
		now := time.Now()
		if _, err = stmt.Exec(now.UnixNano(), now); err != nil {
			panic(err)
		}
	}
}

func (g *Generator) StartUpdate(wg *sync.WaitGroup) {
	defer wg.Done()
	//stmt, err := g.conn.PrepareContext(context.Background(), "update ts_test set ts=current_timestamp(6) where id=1")
	//stmt, err := g.conn.PrepareContext(context.Background(), "update ts_test set ts=? where id=1")
	stmt, err := g.conn.PrepareContext(context.Background(), "update ts_test set ts2=? where id=1")
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			//if _, err = stmt.Exec(); err != nil {
			if _, err = stmt.Exec(time.Now()); err != nil {
				panic(err)
			}
		}
	}
}
