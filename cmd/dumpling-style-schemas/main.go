package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/pingcap/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"try-out/pkg/config"
	"try-out/pkg/constants"
)

func main() {
	if err := config.InitConfig(); err != nil {
		fmt.Printf("Failed to parse command arguments: %v\n", err)
		return
	}
	if err := runWrapper(run); err != nil {
		fmt.Printf("Failed to run: %v\n", err)
		os.Exit(1)
	}
}

func runWrapper(run func() error) error {
	start := time.Now()
	err := run()
	if err == nil {
		fmt.Printf("takes: %s\n", time.Since(start))
	}
	return err
}

func run() error {
	cfg := &config.GlobalCfg
	config.GlobalLog.Info("parameters", zap.Stringer("config", cfg))

	stat, err := os.Stat(cfg.OutputDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Annotate(err, "Failed to stat output directory")
		}
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			return errors.Annotate(err, "Failed to create output directory")
		}
	} else if !stat.IsDir() {
		return errors.Errorf("%s is not a directory", cfg.OutputDir)
	}

	// see https://docs.pingcap.com/tidb/stable/dumpling-overview#format-of-exported-files
	tablesPerDB := cfg.Tables / cfg.Databases
	eg := &errgroup.Group{}
	eg.SetLimit(cfg.Threads)
	for i := 0; i < cfg.Databases; i++ {
		i := i
		eg.Go(func() error {
			dbName := fmt.Sprintf("%s%d", cfg.DBPrefix, i)
			dbSchemaFileName := path.Join(cfg.OutputDir, fmt.Sprintf("%s-schema-create.sql", dbName))
			createDBSQL := "CREATE DATABASE IF NOT EXISTS " + dbName + ";\n"
			if err := os.WriteFile(dbSchemaFileName, []byte(createDBSQL), 0644); err != nil {
				return err
			}
			for j := 0; j < tablesPerDB; j++ {
				tableName := fmt.Sprintf("%s%d", cfg.TableName, j)
				createTableSQL := fmt.Sprintf(constants.DefaultCreateTableTemplate, dbName, tableName)
				tblSchemaFileName := path.Join(cfg.OutputDir, fmt.Sprintf("%s.%s-schema.sql", dbName, tableName))
				if err := os.WriteFile(tblSchemaFileName, []byte(createTableSQL), 0644); err != nil {
					return err
				}
			}
			return nil
		})
	}
	return eg.Wait()
}
