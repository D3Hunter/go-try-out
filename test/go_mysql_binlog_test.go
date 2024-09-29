package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
)

func TestGoMySQL(t *testing.T) {
	// Create a binlog syncer with a unique server id, the server id must be different from other MySQL's.
	// flavor is mysql or mariadb
	cfg := replication.BinlogSyncerConfig{
		ServerID: 100,
		Flavor:   "mysql",
		Host:     "::1",
		//Host:     "172.16.6.107",
		Port:             3306,
		User:             "root",
		Password:         "123456",
		DisableRetrySync: true,
	}
	//cfg := replication.BinlogSyncerConfig{
	//	ServerID: 100,
	//	Flavor:   "mysql",
	//	Host:     "172.16.6.107",
	//	Port:     3306,
	//	User:     "root",
	//	Password: "",
	//}

	cfg.DumpCommandFlag = replication.BINLOG_SEND_ANNOTATE_ROWS_EVENT

	for i := 1; i <= 1; i++ {
		binlogFileName := fmt.Sprintf("mysql-bin.%06d", i)
		fmt.Printf("================== binlog %s ==================\n", binlogFileName)
		// Start sync with specified binlog file and position
		syncer := replication.NewBinlogSyncer(cfg)
		//streamer, _ := syncer.StartSync(mysql.Position{binlogFileName, 4})
		//streamer, _ := syncer.StartSync(mysql.Position{binlogFileName, 4})
		// an non-exist gtid set
		//set, _ := mysql.ParseMysqlGTIDSet("de278ad0-0000-11e4-9f8e-6edd0ca20947:1-3")
		set, _ := mysql.ParseMysqlGTIDSet("")
		streamer, _ := syncer.StartSyncGTID(set)
		//streamer, _ := syncer.StartSyncGTID(set)

		// or you can start a gtid replication like
		// streamer, _ := syncer.StartSyncGTID(gtidSet)
		// the mysql GTID set likes this "de278ad0-2106-11e4-9f8e-6edd0ca20947:1-2"
		// the mariadb GTID set likes this "0-1-100"

		for i := 0; i < 200; i++ {
			ev, _ := streamer.GetEvent(context.Background())
			// Dump event
			ev.Dump(os.Stdout)
		}
		syncer.Close()
	}
}
