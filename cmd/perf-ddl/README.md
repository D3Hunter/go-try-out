# Note

The table template is fixed right now, it's hardcoded in the code.
```sql
CREATE TABLE %s.%s (
  id int(11) NOT NULL AUTO_INCREMENT,
  k int(11) NOT NULL DEFAULT '0',
  c char(120) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  pad char(60) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
  PRIMARY KEY (id) /*T![clustered_index] CLUSTERED */,
  KEY k_613 (k)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci /*T![auto_id_cache] AUTO_ID_CACHE=1 */
```

# Build

```shell
make perf-ddl
```

# Common used commands for DDL operations

create 100K tables inside a single database `aa` with 10 threads.

```
nohup /perf-ddl --host tc-tidb-0.tc-tidb-peer --database-name aa --action create-tables-on-single-db --tables 100000 --threads 10 > nohup.log 2>&1 &
```

create 100K databases with prefix `bb` with 10 threads.

```
nohup /perf-ddl --host tc-tidb-0.tc-tidb-peer --databases 100000 --db-prefix bb --action create-databases --threads 10 > nohup.log 2>&1 &
```

create 10K tenants, where each tenant is a DB of prefix `tenant` with 100 tables, with 10 threads.

```
```sql
nohup /perf-ddl --host tc-tidb-0.tc-tidb-peer --action init-multiple-tenants --db-prefix tenant --databases 10000 --tables 1000000 --threads 10 > nohup.log 2>&1 &
```

do table level DDL on 100K tables with 10 threads and a DDL template.

```
nohup /perf-ddl --host tc-tidb-0.tc-tidb-peer --database-name aa --action table-level-ddl --table-ddl-template 'alter table %s.%s add column v int' --tables 100000 --threads 10 > nohup.log 2>&1 &
```