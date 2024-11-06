package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"try-out/pkg/config"
	"try-out/pkg/tidb"
)

type TableSQL struct {
	Table       string
	SQLTemplate string
}

// run with:
//
//	nohup /customer-create --host tc-tidb-0.tc-tidb-peer --threads 10 --sql-dir /ddls_creation > nohup.log 2>&1 &
//
// the sql-dir should have this structure:
//
//	 .
//	 ├── schema1
//	 │   ├── schema1-table1-schema.sql
//	 │   └── schema1-table2-schema.sql
//	 └── schema2
//		├── schema2-table1-schema.sql
//		└── schema2-table2-schema.sql
//
// and the table-name inside the sql file should be enclosed in backticks, like:
//
//	CREATE TABLE `table1` (
//	  id int(11) NOT NULL AUTO_INCREMENT,
//	  k int(11) NOT NULL DEFAULT '0'
//	)
func main() {
	if err := config.InitConfig(); err != nil {
		fmt.Printf("Failed to parse command arguments: %v\n", err)
		return
	}
	config.GlobalLog.Info("parameters", zap.String("host", config.GlobalCfg.Host), zap.Int("port", config.GlobalCfg.Port), zap.Int("threads", config.GlobalCfg.Threads),
		zap.Int("database", config.GlobalCfg.Databases), zap.Int("tables", config.GlobalCfg.Tables), zap.String("action", config.GlobalCfg.Action))

	allSQLs := make(map[string][]TableSQL, 4)
	dirFS := os.DirFS(config.GlobalCfg.SQLDir)
	if err := fs.WalkDir(dirFS, ".", func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) != ".sql" {
			return nil
		}
		content, err := os.ReadFile(filepath.Join(config.GlobalCfg.SQLDir, path))
		if err != nil {
			return err
		}
		parts := strings.Split(path, fmt.Sprintf("%c", filepath.Separator))
		fileName := parts[len(parts)-1]
		schema := parts[len(parts)-2]
		tableName := strings.TrimPrefix(fileName, schema+".")
		tableName = strings.TrimSuffix(tableName, "-schema.sql")
		allSQLs[schema] = append(allSQLs[schema], TableSQL{
			Table:       tableName,
			SQLTemplate: string(content),
		})
		return nil
	}); err != nil {
		fmt.Printf("Failed to walk dir: %v\n", err)
		return
	}
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.GlobalCfg.User, config.GlobalCfg.Password,
		config.GlobalCfg.Host, config.GlobalCfg.Port,
	))
	if err != nil {
		fmt.Printf("Failed to connect to MySQL database: %v\n", err)
		return
	}
	defer db.Close()
	ctx := context.Background()
	for schema := range allSQLs {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", schema)); err != nil {
			fmt.Printf("Failed to create database %s: %v\n", schema, err)
			return
		}
	}
	if err := tidb.ShowVersion(db); err != nil {
		fmt.Printf("Failed to show version: %v\n", err)
		return
	}

	conns, err := tidb.PrepareConnections(db, config.GlobalCfg.Threads)
	if err != nil {
		fmt.Printf("Failed to prepare connections: %v\n", err)
		return
	}
	for schema, tableSQLs := range allSQLs {
		for _, tableSQL := range tableSQLs {
			config.GlobalLog.Info("create tables using template",
				zap.String("schema", schema),
				zap.String("base-table-name", tableSQL.Table))
			if err := createTablesUsingSQLTemplate(conns, schema, tableSQL); err != nil {
				fmt.Printf("Failed to create tables: %v\n", err)
				return
			}
		}
	}
}

func createTablesUsingSQLTemplate(conns []*sql.Conn, schema string, tableSQL TableSQL) error {
	ctx := context.Background()
	var eg errgroup.Group
	var idAlloc atomic.Int32
	countPerThread := (40 * 500) / config.GlobalCfg.Threads
	for i := 0; i < config.GlobalCfg.Threads; i++ {
		conn := conns[i]
		eg.Go(func() error {
			for j := 0; j < countPerThread; j++ {
				newFullTableName := fmt.Sprintf("%s.%s_%d", schema, tableSQL.Table, idAlloc.Inc())
				newSQL := strings.Replace(tableSQL.SQLTemplate, fmt.Sprintf("`%s`", tableSQL.Table), newFullTableName, 1)
				if _, err := conn.ExecContext(ctx, newSQL); err != nil {
					fmt.Printf("Failed to create table %s: %v\n", tableSQL.Table, err)
					return err
				}
			}
			return nil
		})
	}
	return eg.Wait()
}
