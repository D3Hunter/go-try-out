package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/sstable"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestPebble(t *testing.T) {
	//writePebble()

	testWriteSST()

	testReadSST()
}

func testWriteSST() {
	file, err := os.OpenFile("demo/test.sst", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	writer := sstable.NewWriter(file, sstable.WriterOptions{})
	defer writer.Close()
	for i := 0; i < 3; i++ {
		err := writer.Set([]byte("hel"), []byte("world-"+strconv.Itoa(i)))
		if err != nil {
			return
		}
	}
	for i := 0; i < 3; i++ {
		err := writer.Set([]byte("hello"), []byte("world-"+strconv.Itoa(i)))
		if err != nil {
			return
		}
	}
}

func testReadSST() {
	file, err := os.Open("demo/test.sst")
	if err != nil {
		panic(err)
	}
	reader, err := sstable.NewReader(file, sstable.ReaderOptions{})
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	iter, err := reader.NewIter(nil, nil)
	if err != nil {
		panic(err)
	}
	defer iter.Close()
	for key, bytes := iter.Next(); key != nil; key, bytes = iter.Next() {
		fmt.Println(key, string(bytes))
	}
}

func writePebble() {
	db, err := pebble.Open("demo", &pebble.Options{})
	if err != nil {
		log.Fatal(err)
	}
	key := []byte("hello")
	for i := 0; i < 3; i++ {
		if err := db.Set(key, []byte("world"), pebble.Sync); err != nil {
			log.Fatal(err)
		}
	}
	value, closer, err := db.Get(key)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s %s\n", key, value)
	if err := closer.Close(); err != nil {
		log.Fatal(err)
	}
	db.Flush()
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}

func Test(t *testing.T) {
	fmt.Println("before")
	require.True(t, false)
	defer fmt.Println("never run")
}

func TestQueryMySQL(t *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:4000)/dataflow") //?parseTime=true&loc=Asia%2FShanghai
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
	queryAndPrint(t, conn, "select version()")
	executePsAndPrint(t, conn, "select * from t WHERE TIMESTAMPDIFF(SECOND, updated_at, UTC_TIMESTAMP()) > ?", 300.0)
	queryAndPrint(t, conn, "select * from t WHERE TIMESTAMPDIFF(SECOND, updated_at, UTC_TIMESTAMP()) > 300.000000")
	//executePsAndPrint(t, conn, "select * from t2 WHERE id > ?", 1)
	//queryAndPrint(t, conn, "select * from t2 WHERE id > 1.000000")
}

func TestQueryMySQLGorm(t *testing.T) {
	gormConfig := &gorm.Config{}

	db, err := gorm.Open(gormmysql.Open("root:123456@tcp(127.0.0.1:4000)/dataflow"), gormConfig)
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	queryAndPrint(t, conn, "select version()")
	queryAndPrint(t, conn, "SELECT * FROM `df_private_link_endpoints` WHERE TIMESTAMPDIFF(SECOND, updated_at, UTC_TIMESTAMP()) > 300.000000 AND phase = 'failed';")
}

func Test_insert_mysql(t *testing.T) {
	//db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3307)/") //?parseTime=true&loc=Asia%2FShanghai
	//db, err := sql.Open("mysql", "root:@tcp(172.16.102.104:4000)/test") //?parseTime=true&loc=Asia%2FShanghai
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:4100)/test") //?parseTime=true&loc=Asia%2FShanghai
	//db, err := sql.Open("mysql", "root:12345678@tcp(tidb.ca17169f.1db98375.ap-southeast-1.prod.aws.tidbcloud.com:4000)/") //?parseTime=true&loc=Asia%2FShanghai
	if err != nil {
		panic(err)
	}
	defer db.Close()
	for i := 0; i < 10000000000; i++ {
		var j int
		for {
			start := time.Now()
			_, err = db.ExecContext(context.Background(), "insert into foo values(1)")
			err2 := errors.Cause(err)
			switch nerr := err2.(type) {
			case net.Error:
				if nerr.Timeout() {
					fmt.Println("timeout")
				}
				switch cause := nerr.(type) {
				case *net.OpError:
					syscallErr, ok := cause.Unwrap().(*os.SyscallError)
					if ok {
						fmt.Println(syscallErr.Err == syscall.ECONNREFUSED || syscallErr.Err == syscall.ECONNRESET)
					}
				}
			}
			if err2 != nil {
				fmt.Println("db failed", i, j, err2, time.Since(start).Milliseconds())
				//time.Sleep(100 * time.Millisecond)
			} else {
				fmt.Println("db success", i, j, time.Since(start).Milliseconds())
				break
			}
			time.Sleep(5 * time.Millisecond)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestScanNil(t *testing.T) {
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/test") //?parseTime=true&loc=Asia%2FShanghai
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
	rows, err := conn.QueryContext(ctx, "select v, v1 from t limit 1")
	if err != nil {
		panic(err)
	}
	colVals := make([][]byte, 2)
	colValsI := make([]interface{}, len(colVals))
	for i := range colValsI {
		colValsI[i] = &colVals[i]
	}

	rows.Next()
	err = rows.Scan(colValsI...)
	if err != nil {
		panic(err)
	}
	fmt.Println(colValsI)
}

func Test_get_mysql_warning(t *testing.T) {
	db, err := sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/test?sql_mode=''") //?parseTime=true&loc=Asia%2FShanghai
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
	result, err := conn.ExecContext(ctx, "insert into t values('2,12')")
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func queryAndPrint(t *testing.T, conn *sql.Conn, sql string) {
	fmt.Println("execute using stmt:")
	rows, err := conn.QueryContext(context.Background(), sql)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	printRows(t, rows)
}

func printRows(t *testing.T, rows *sql.Rows) {
	columns, err := rows.Columns()
	require.NoError(t, err)
	dst := make([]interface{}, len(columns))
	for i := range columns {
		dst[i] = &[]byte{}
	}
	for rows.Next() {
		err := rows.Scan(dst...)
		require.NoError(t, err)
		var s string
		for _, v := range dst {
			s += fmt.Sprintf("%v, ", string(*v.(*[]byte)))
		}
		fmt.Println("Row:", s)
	}
}

func executePsAndPrint(t *testing.T, conn *sql.Conn, sql string, args ...any) {
	fmt.Println("execute using prepared stmt:")
	ctx := context.Background()
	ps, err2 := conn.PrepareContext(ctx, sql)
	require.NoError(t, err2)
	rows, err2 := ps.QueryContext(ctx, args...)
	require.NoError(t, err2)
	defer rows.Close()
	printRows(t, rows)
}

func TestSha256(t *testing.T) {
	sum := sha256.Sum256([]byte("hello world\n"))
	fmt.Printf("%x\n", sum)
	if t == nil {
		sum := 0
		_ = sum
	}
	m := make(map[int]int)
	m[1] = 100
	m[2] = 300
	m[3] = 222
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	logger.Info("map", zap.Reflect("map", m))
	s := make([]int64, 3)
	s[2] = 100
	logger.Info("slice", zap.Int64s("s", s))
}

func TestPreparedSQLWithColumnQuestion(t *testing.T) {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:4000)/test") //?parseTime=true&loc=Asia%2FShanghai
	require.NoError(t, err)
	defer func() {
		require.NoError(t, db.Close())
	}()
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()
	stmt, err := conn.PrepareContext(ctx, "select id from t where 1 = ?")
	require.NoError(t, err)
	rows, err := stmt.Query("id")
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var id int
		require.NoError(t, rows.Scan(&id))
		fmt.Println(id)
	}
	require.NoError(t, rows.Err())
}

func TestJoin(t *testing.T) {
	fmt.Println(path.Join("a", ""))
}

func TestRateLimiter(t *testing.T) {
	limiter := rate.NewLimiter(rate.Every(time.Second), 1)
	for i := 0; i < 1000; i++ {
		if limiter.Allow() {
			fmt.Printf("allow: %s\n", time.Now())
		}
		time.Sleep(100 * time.Millisecond)
	}
}
