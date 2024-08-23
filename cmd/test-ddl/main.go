package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
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

type config struct {
	host    string
	port    int
	user    string
	threads int
	// total number of databases to create
	databases int
	// total number of tables to create
	tables              int
	action              string
	dbPrefix            string
	databaseName        string
	tableDDLTemplate    string
	skipPrepare         bool
	analyzeLog          bool
	logFile             string
	analyzeStartTimeStr string
	analyzeEndTimeStr   string
}

var (
	globalCfg config
	globalLog = zap.NewExample()
)

func initConfig() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&globalCfg.host, "host", "tc-tidb", "host")
	fs.IntVar(&globalCfg.port, "port", 4000, "port")
	fs.StringVar(&globalCfg.user, "user", "root", "user")
	fs.IntVar(&globalCfg.threads, "threads", 8, "threads")
	fs.IntVar(&globalCfg.databases, "databases", 1, "number of databases")
	fs.IntVar(&globalCfg.tables, "tables", 1, "total number of tables")
	fs.StringVar(&globalCfg.action, "action", "create-tables-on-single-db", "action")
	fs.StringVar(&globalCfg.dbPrefix, "db-prefix", "db", "database prefix")
	fs.StringVar(&globalCfg.databaseName, "database-name", "db", "database name")
	fs.StringVar(&globalCfg.tableDDLTemplate, "table-ddl-template", "", "table ddl template")
	fs.BoolVar(&globalCfg.skipPrepare, "skip-prepare", false, "skip prepare")
	fs.BoolVar(&globalCfg.analyzeLog, "analyze-log", false, "analyze log")
	fs.StringVar(&globalCfg.logFile, "log-file", "", "log file")
	fs.StringVar(&globalCfg.analyzeStartTimeStr, "analyze-start-time", "", "analyze start time")
	fs.StringVar(&globalCfg.analyzeEndTimeStr, "analyze-end-time", "", "analyze end time")

	return fs.Parse(os.Args[1:])
}

type result struct {
	startTime time.Time
}

func main() {
	if err := initConfig(); err != nil {
		fmt.Printf("Failed to parse command arguments: %v\n", err)
		return
	}
	globalLog.Info("parameters", zap.String("host", globalCfg.host), zap.Int("port", globalCfg.port), zap.Int("threads", globalCfg.threads),
		zap.Int("database", globalCfg.databases), zap.Int("tables", globalCfg.tables), zap.String("action", globalCfg.action))

	if globalCfg.action == "analyze-log" {
		time.Sleep(time.Second) // wait all logs flushed
		if globalCfg.logFile == "" {
			fmt.Println("log file is required for analyze-log")
			return
		}
		analyzeStartTime, err := time.Parse(logTimeFormat, globalCfg.analyzeStartTimeStr)
		if err != nil {
			fmt.Println("failed to parse execute start time: ", err)
			return
		}
		var analyzeEndTime time.Time
		if globalCfg.analyzeEndTimeStr != "" {
			analyzeEndTime, err = time.Parse(logTimeFormat, globalCfg.analyzeEndTimeStr)
			if err != nil {
				fmt.Println("failed to parse execute end time: ", err)
				return
			}
		}
		if err := analyzeCallCost(analyzeStartTime, analyzeEndTime, globalCfg.logFile); err != nil {
			fmt.Printf("Failed to analyze log: %v\n", err)
		}
		return
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s@tcp(%s:%d)/", globalCfg.user, globalCfg.host, globalCfg.port))
	if err != nil {
		fmt.Printf("Failed to connect to MySQL database: %v\n", err)
		return
	}
	defer db.Close()
	if err := showVersion(db); err != nil {
		fmt.Printf("Failed to show version: %v\n", err)
		return
	}

	var res result
	switch globalCfg.action {
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

	if globalCfg.analyzeLog {
		time.Sleep(time.Second)
		if globalCfg.logFile == "" {
			fmt.Println("log file is required for analyze-log")
			return
		}
		if err := analyzeCallCost(res.startTime, time.Time{}, globalCfg.logFile); err != nil {
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
