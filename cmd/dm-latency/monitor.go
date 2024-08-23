package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

/*
	create table ts_test(id bigint AUTO_INCREMENT primary key, ts bigint, ts2 timestamp(6));
	insert into ts_test(ts, ts2) values(0, current_timestamp(6));

	alter table ts_test add column updated_at timestamp(6) default CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6);

	create table ts_test(id bigint AUTO_INCREMENT primary key, ts timestamp(6));
	insert into ts_test(ts) values(current_timestamp(6));
*/

type Monitor struct {
	sourceConn   *sql.Conn
	targetConn   *sql.Conn
	warmingUpCnt int32
	cnt          int64
	diffSum      int64
}

func NewMonitor() *Monitor {
	return &Monitor{warmingUpCnt: 5}
}

func (m *Monitor) Init() error {
	var err error
	if m.sourceConn, err = OpenDbConn("root:123456@tcp(127.0.0.1:3306)/test?loc=Local"); err != nil {
		return err
	}
	if m.targetConn, err = OpenDbConn("root:@tcp(127.0.0.1:4000)/test"); err != nil {
		return err
	}
	return nil
}

func (m *Monitor) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var cnt, diffSum int64
	for {
		select {
		case <-ticker.C:
			diff, err := m.getTsDiffUsingUpdateTime()
			if err != nil {
				continue
			}
			if m.warmingUpCnt > 0 {
				m.warmingUpCnt--
				continue
			}

			cnt++
			diffSum += diff
			m.cnt++
			m.diffSum += diff
		}

		if cnt%10 == 0 {
			fmt.Printf("current migration latency: %d ms, accumulated average: %d ms\n",
				time.Duration(diffSum/cnt).Milliseconds(), time.Duration(m.diffSum/m.cnt).Milliseconds())
			cnt = 0
			diffSum = 0
		}
	}
}

func (m *Monitor) getTsDiffUsingUpdateTime() (int64, error) {
	row := m.targetConn.QueryRowContext(context.Background(), "select timestampdiff(MICROSECOND, ts2, updated_at) from ts_test")
	var val float64
	if err := row.Scan(&val); err != nil {
		fmt.Println("failed to get max ts from source.")
		return 0, err
	}
	return int64(val * 1000), nil
}

func (m *Monitor) getTsDiff() (int64, error) {
	sourceTs, err := m.getMaxTs(m.sourceConn)
	if err != nil {
		return 0, err
	}
	targetTs, err := m.getMaxTs(m.targetConn)
	if err != nil {
		return 0, err
	}
	diff := sourceTs - targetTs
	if diff > 0 {
		return diff, nil
	} else {
		return 0, nil
	}
}

func (m *Monitor) getMaxTs(conn *sql.Conn) (int64, error) {
	//row := conn.QueryRowContext(context.Background(), "select ts from ts_test where id=1")
	//var ts int64
	//if err := row.Scan(&ts); err != nil {
	//	fmt.Println("failed to get max ts from source.")
	//	return 0, err
	//}
	//return ts, nil

	row := conn.QueryRowContext(context.Background(), "select max(ts2) from ts_test")
	var tsStr string
	if err := row.Scan(&tsStr); err != nil {
		fmt.Println("failed to get max ts from source.")
		return 0, err
	}
	ts, err := time.ParseInLocation("2006-01-02 15:04:05.000000", tsStr, time.Local)
	if err != nil {
		return 0, err
	}
	return ts.UnixNano(), nil
}
