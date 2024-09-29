
perf-ddl:
	go build -o bin/perf-ddl cmd/perf-ddl/*.go
	GOOS=linux GOARCH=amd64 go build -o bin/perf-ddl-linux-amd64 cmd/perf-ddl/*.go

write-hotspot:
	go build -o bin/write-hotspot cmd/write-hotspot/*.go
	GOOS=linux GOARCH=amd64 go build -o bin/write-hotspot-linux-amd64 cmd/write-hotspot/*.go

customer-create:
	go build -o bin/customer-create cmd/customer-create/*.go
	GOOS=linux GOARCH=amd64 go build -o bin/customer-create-linux-amd64 cmd/customer-create/*.go

github-crawl:
	go build -o bin/github-crawl cmd/github-crawl/*.go

dumpling-style-schemas:
	go build -o bin/dumpling-style-schemas cmd/dumpling-style-schemas/*.go
	GOOS=linux GOARCH=amd64 go build -o bin/dumpling-style-schemas-linux-amd64 cmd/dumpling-style-schemas/*.go

gc-cpu-usage:
	go build -o bin/gc-cpu-usage cmd/gc-cpu-usage/*.go
	GOOS=linux GOARCH=amd64 go build -o bin/gc-cpu-usage-linux-amd64 cmd/gc-cpu-usage/*.go

bazel-prepare:
	bazel run //:gazelle
	bazel run //:gazelle-update-repos