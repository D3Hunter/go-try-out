package config

import (
	"encoding/json"
	"flag"
	"os"

	"go.uber.org/zap"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Threads  int
	// total number of databases to create
	Databases int
	// total number of tables to create
	Tables              int
	Action              string
	DBPrefix            string
	DatabaseName        string
	TableName           string
	TableDDLTemplate    string
	SkipPrepare         bool
	AnalyzeLog          bool
	LogFile             string
	AnalyzeStartTimeStr string
	AnalyzeEndTimeStr   string
	SQLDir              string
	OutputDir           string
}

func (c *Config) String() string {
	contents, _ := json.Marshal(c)
	return string(contents)
}

var (
	GlobalCfg Config
	GlobalLog = zap.NewExample()
)

func InitConfig() error {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&GlobalCfg.Host, "host", "localhost", "host")
	fs.IntVar(&GlobalCfg.Port, "port", 4000, "port")
	fs.StringVar(&GlobalCfg.User, "user", "root", "user")
	fs.StringVar(&GlobalCfg.Password, "password", "", "password")
	fs.IntVar(&GlobalCfg.Threads, "threads", 8, "threads")
	fs.IntVar(&GlobalCfg.Databases, "databases", 1, "number of databases")
	fs.IntVar(&GlobalCfg.Tables, "tables", 1, "total number of tables")
	fs.StringVar(&GlobalCfg.Action, "action", "create-tables-on-single-db", "action")
	fs.StringVar(&GlobalCfg.DBPrefix, "db-prefix", "db", "database prefix")
	fs.StringVar(&GlobalCfg.DatabaseName, "database-name", "db", "database name")
	fs.StringVar(&GlobalCfg.TableName, "table-name", "t", "table name")
	fs.StringVar(&GlobalCfg.TableDDLTemplate, "table-ddl-template", "", "table ddl template")
	fs.BoolVar(&GlobalCfg.SkipPrepare, "skip-prepare", false, "skip prepare")
	fs.BoolVar(&GlobalCfg.AnalyzeLog, "analyze-log", false, "analyze log")
	fs.StringVar(&GlobalCfg.LogFile, "log-file", "", "log file")
	fs.StringVar(&GlobalCfg.AnalyzeStartTimeStr, "analyze-start-time", "", "analyze start time")
	fs.StringVar(&GlobalCfg.AnalyzeEndTimeStr, "analyze-end-time", "", "analyze end time")
	fs.StringVar(&GlobalCfg.SQLDir, "sql-dir", "", "sql dir")
	fs.StringVar(&GlobalCfg.OutputDir, "output-dir", "dumps", "output dir")

	return fs.Parse(os.Args[1:])
}
