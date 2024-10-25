package main

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"try-out/pkg/config"
	"try-out/pkg/tidb"
)

var (
	TableSQL = `
CREATE TABLE %s.%s (
  id int(11) NOT NULL AUTO_INCREMENT,
  k int(11) NOT NULL DEFAULT '0',
  c char(120) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  pad char(60) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  PRIMARY KEY (id) /*T![clustered_index] CLUSTERED */,
  KEY k_613 (k)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci /*T![auto_id_cache] AUTO_ID_CACHE=1 */
`
)

type result struct {
	startTime time.Time
}

func main() {
	if err := config.InitConfig(); err != nil {
		fmt.Printf("Failed to parse command arguments: %v\n", err)
		return
	}
	config.GlobalLog.Info("parameters", zap.String("host", config.GlobalCfg.Host), zap.Int("port", config.GlobalCfg.Port), zap.Int("threads", config.GlobalCfg.Threads),
		zap.Int("database", config.GlobalCfg.Databases), zap.Int("tables", config.GlobalCfg.Tables), zap.String("action", config.GlobalCfg.Action))

	if config.GlobalCfg.Action == "analyze-log" {
		time.Sleep(time.Second) // wait all logs flushed
		if config.GlobalCfg.LogFile == "" {
			fmt.Println("log file is required for analyze-log")
			return
		}
		analyzeStartTime, err := time.Parse(logTimeFormat, config.GlobalCfg.AnalyzeStartTimeStr)
		if err != nil {
			fmt.Println("failed to parse execute start time: ", err)
			return
		}
		var analyzeEndTime time.Time
		if config.GlobalCfg.AnalyzeEndTimeStr != "" {
			analyzeEndTime, err = time.Parse(logTimeFormat, config.GlobalCfg.AnalyzeEndTimeStr)
			if err != nil {
				fmt.Println("failed to parse execute end time: ", err)
				return
			}
		}
		if err := analyzeCallCost(analyzeStartTime, analyzeEndTime, config.GlobalCfg.LogFile); err != nil {
			fmt.Printf("Failed to analyze log: %v\n", err)
		}
		return
	}

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

	var res result
	switch config.GlobalCfg.Action {
	case "create-tables-on-single-db":
		res = createTablesOnSingleDBAction(db)
	case "create-databases":
		res = createDatabasesAction(db)
	case "init-multiple-tenants":
		res = testInitMultipleTenantsAction(db)
	case "table-level-ddl":
		// tables must be created using "create-tables-on-single-db" action
		res = tableLevelDDLAction(db)
	}

	if config.GlobalCfg.AnalyzeLog {
		time.Sleep(time.Second)
		if config.GlobalCfg.LogFile == "" {
			fmt.Println("log file is required for analyze-log")
			return
		}
		if err := analyzeCallCost(res.startTime, time.Time{}, config.GlobalCfg.LogFile); err != nil {
			fmt.Printf("Failed to analyze log: %v\n", err)
		}
		return
	}
}

func printPercentile(prefix string, durations []time.Duration) {
	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})
	sum := time.Duration(0)
	for _, d := range durations {
		sum += d
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: count: %d, avg: %v", prefix, len(durations),
		(sum / time.Duration(len(durations))).Round(time.Millisecond)))
	for i := 0; i <= 10; i++ {
		idx := int(float64(len(durations)) * float64(i*10) / 100)
		if idx >= len(durations) {
			idx = len(durations) - 1
		}
		key := fmt.Sprintf("P%d", i*10)
		if i == 0 {
			key = "min"
		} else if i == 10 {
			key = "max"
		}
		sb.WriteString(fmt.Sprintf(", %s: %v", key, durations[idx].Round(time.Millisecond)))
	}
	fmt.Println(sb.String())
}
