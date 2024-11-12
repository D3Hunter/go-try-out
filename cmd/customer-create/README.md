# Note
This program will run every table 200K times, it's hardcoded right now.

The sql-dir should have this structure:
```
 .
 ├── schema1
 │   ├── schema1-table1-schema.sql
 │   └── schema1-table2-schema.sql
 └── schema2
	├── schema2-table1-schema.sql
	└── schema2-table2-schema.sql
```
and the table-name inside the sql file should be enclosed in backticks, like:
```sql
CREATE TABLE `table1` (
  id int(11) NOT NULL AUTO_INCREMENT,
  k int(11) NOT NULL DEFAULT '0'
)
```

# Build

```shell
make customer-create
```

# Common used commands for DDL operations

create schema and tables inside `/ddls_creation` directory with 10 threads.

```
nohup /customer-create --host tc-tidb-0.tc-tidb-peer --threads 10 --sql-dir /ddls_creation > nohup.log 2>&1 &
```
