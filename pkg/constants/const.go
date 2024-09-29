package constants

var (
	GlobalVar = "global"

	// DefaultCreateTableTemplate is the default create table template, it's same
	// as the "sbtest" table of sysbench.
	// takes about 2.4k to store this table in TiDB info schema cache.
	DefaultCreateTableTemplate = `
CREATE TABLE %s.%s (
  id int(11) NOT NULL AUTO_INCREMENT,
  k int(11) NOT NULL DEFAULT '0',
  c char(120) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  pad char(60) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  PRIMARY KEY (id) /*T![clustered_index] CLUSTERED */,
  KEY k_613 (k)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci /*T![auto_id_cache] AUTO_ID_CACHE=1 */;
`
	// CreateTableTemplate8K takes about 8K to store this table in TiDB info schema cache.
	CreateTableTemplate8K = `
CREATE TABLE %s.%s (
  id bigint(20) NOT NULL AUTO_INCREMENT,
  column_1 int(11) DEFAULT NULL COMMENT 'this is column_1',
  column_2 varchar(255) DEFAULT NULL COMMENT 'this is column_2',
  column_3 varchar(255) DEFAULT NULL COMMENT 'this is column_3',
  column_4 int(11) DEFAULT NULL COMMENT 'this is column_4',
  column_5 int(11) DEFAULT NULL COMMENT 'this is column_5',
  column_6 datetime DEFAULT NULL COMMENT 'this is column_6',
  column_7 datetime DEFAULT NULL COMMENT 'this is column_7',
  column_8 text COMMENT 'this is column_8',
  column_9 mediumint(8) unsigned NOT NULL DEFAULT '0' COMMENT 'this is column_9',
  column_10 varchar(255) DEFAULT NULL COMMENT 'this is column_10',
  column_11 int(11) DEFAULT NULL COMMENT 'this is column_11',
  column_12 int(10) unsigned NOT NULL DEFAULT '0' COMMENT 'this is column_12',
  column_13 datetime DEFAULT NULL COMMENT 'this is column_13',
  column_14 datetime DEFAULT NULL COMMENT 'this is column_14',
  column_15 datetime DEFAULT NULL COMMENT 'this is column_15',
  column_16 datetime DEFAULT NULL COMMENT 'this is column_16',
  PRIMARY KEY (id),
  INDEX idx_column_1(column_1),
  INDEX idx_column_2(column_2),
  INDEX idx_column_3(column_3),
  INDEX idx_column_4(column_4)
) ENGINE=InnoDB AUTO_INCREMENT=209 DEFAULT CHARSET utf8mb4 COLLATE utf8mb4_general_ci COMMENT 'this is the table';
`
	ParentTableTemplate = `
CREATE TABLE %s.%s (
  id int(11) NOT NULL AUTO_INCREMENT,
  k int(11) NOT NULL DEFAULT '0',
  c char(120) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  pad char(60) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  PRIMARY KEY (id) /*T![clustered_index] CLUSTERED */,
  KEY k_613 (k)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
`
	ChildTableTemplate = `
CREATE TABLE %s.%s (
  id int(11) NOT NULL AUTO_INCREMENT,
  k int(11) NOT NULL DEFAULT '0',
  c char(120) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  pad char(60) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  PRIMARY KEY (id) /*T![clustered_index] CLUSTERED */,
  KEY k_613 (k),
  CONSTRAINT fk_parent_id FOREIGN KEY (id) REFERENCES %s.%s (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
`
)
