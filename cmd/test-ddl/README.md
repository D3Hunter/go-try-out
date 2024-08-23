# Common used commands for DDL operations

create 100K tables inside a single database `aa` with 10 threads.

```
nohup /test-ddl --host tc-tidb-0.tc-tidb-peer --database-name aa --action create-tables-on-single-db --tables 100000 --threads 10 > nohup.log 2>&1 &
```

create 100K databases with prefix `bb` with 10 threads.

```
nohup /test-ddl --host tc-tidb-0.tc-tidb-peer --databases 100000 --db-prefix bb --action create-databases --threads 10 > nohup.log 2>&1 &
```

create 10K tenants, where each tenant is a DB of prefix `tenant` with 100 tables, with 10 threads.

```
```sql
nohup /test-ddl --host tc-tidb-0.tc-tidb-peer --action init-multiple-tenants --db-prefix tenant --databases 10000 --tables 1000000 --threads 10 > nohup.log 2>&1 &
```

do table level DDL on 100K tables with 10 threads and a DDL template.

```
nohup /test-ddl --host tc-tidb-0.tc-tidb-peer --database-name aa --action table-level-ddl --table-ddl-template 'alter table %s.%s add column v int' --tables 100000 --threads 10 > nohup.log 2>&1 &
```