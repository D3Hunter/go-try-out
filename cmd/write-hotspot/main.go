package main

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"try-out/pkg/config"
	"try-out/pkg/tidb"
)

func main() {
	if err := config.InitConfig(); err != nil {
		fmt.Printf("Failed to parse command arguments: %v\n", err)
		return
	}
	config.GlobalLog.Info("parameters", zap.String("host", config.GlobalCfg.Host), zap.Int("port", config.GlobalCfg.Port), zap.Int("threads", config.GlobalCfg.Threads),
		zap.Int("database", config.GlobalCfg.Databases), zap.Int("tables", config.GlobalCfg.Tables), zap.String("action", config.GlobalCfg.Action))

	db, err := sql.Open("mysql", fmt.Sprintf("%s@tcp(%s:%d)/", config.GlobalCfg.User, config.GlobalCfg.Host, config.GlobalCfg.Port))
	if err != nil {
		fmt.Printf("Failed to connect to MySQL database: %v\n", err)
		return
	}
	defer db.Close()
	if err := tidb.ShowVersion(db); err != nil {
		fmt.Printf("Failed to show version: %v\n", err)
		return
	}

	conns, err := tidb.PrepareConnections(db, config.GlobalCfg.Threads)
	if err != nil {
		fmt.Printf("Failed to prepare connections: %v\n", err)
		return
	}
	defer tidb.RecycleConnections(conns)
	wg := sync.WaitGroup{}
	wg.Add(config.GlobalCfg.Threads)
	batchSize := 8
	baseSeed := time.Now().UnixNano()
	total := int64(batchSize * config.GlobalCfg.Threads * (1 << 20))
	maxID := total * 50
	config.GlobalLog.Info("start",
		zap.Int64("baseSeed", baseSeed),
		zap.Int64("total", total),
		zap.Int64("maxID", maxID))
	//for i := 0; i < config.GlobalCfg.Threads; i++ {
	//	conn := conns[i]
	//	rnd := rand.New(rand.NewSource(baseSeed + int64(i)))
	//	go func() {
	//		defer wg.Done()
	//		batchInsertWithRetryLoop(db, conn, rnd, maxID, batchSize)
	//	}()
	//}
	wg.Wait()
}

type inserter struct {
	db        *sql.DB
	conn      *sql.Conn
	rnd       *rand.Rand
	maxID     int64
	batchSize int
}

func (is *inserter) batchInsertWithRetryLoop() {
	longStr := getLongStr()
	for i := 0; i < 1<<20; i++ {
		rows := make([]string, 0, is.batchSize)
		for j := 0; j < is.batchSize; j++ {
			id := is.rnd.Int63n(is.maxID)
			rows = append(rows, fmt.Sprintf("(%d, %d, %d, %d, '%s')", id, id, id, id, longStr))
		}
		if err := is.batchInsertWithRetry(rows); err != nil {
			is.tryResetConnAsNeeded(err)
		}
	}
}

func (is *inserter) tryResetConnAsNeeded(err error) {
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "invalid connection") ||
		strings.Contains(errStr, "bad connection") ||
		strings.Contains(errStr, "connection is already closed") {
		_ = is.conn.Close()
		is.conn, err = is.db.Conn(context.Background())
		if err != nil {
			config.GlobalLog.Error("failed to refresh conn", zap.Error(err))
		}
	}
}

func (is *inserter) batchInsertWithRetry(rows []string) error {
	template := fmt.Sprintf("insert into test.%s(id, a, b, c, d) values", config.GlobalCfg.TableName)
	err := execSql(is.conn, template+strings.Join(rows, ","))
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return err
		}
		is.tryResetConnAsNeeded(err)
		for _, r := range rows {
			for i := 0; i < 2; i++ {
				err2 := execSql(is.conn, template+r)
				if err2 == nil || strings.Contains(err2.Error(), "Duplicate entry") {
					break
				}
				fmt.Printf("Failed to insert data: %v\n", err2)
				is.tryResetConnAsNeeded(err2)
			}
		}
	}
	return nil
}

func execSql(conn *sql.Conn, sql string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := conn.ExecContext(ctx, sql)
	return err
}

func insertLoop(conn *sql.Conn) {
	data := getLongStr()
	for j := 0; j < 1<<20; j++ {
		_, err := conn.ExecContext(context.Background(), fmt.Sprintf(`insert into test.%s(a,b,c,d) values(1,2,3, '%s')`, config.GlobalCfg.TableName, data))
		if err != nil {
			fmt.Printf("Failed to insert data: %v\n", err)
		}
	}
}

func getLongStr() string {
	data := make([]byte, 5<<9)
	for i := 0; i < len(data); i++ {
		data[i] = 'a'
	}
	return string(data)
}
