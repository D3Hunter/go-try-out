
test-ddl:
	go build -o bin/test-ddl cmd/test-ddl/*.go
	GOOS=linux GOARCH=amd64 go build -o bin/test-ddl-linux-amd64 cmd/test-ddl/*.go

write-hotspot:
	go build -o bin/write-hotspot cmd/write-hotspot/*.go
	GOOS=linux GOARCH=amd64 go build -o bin/write-hotspot-linux-amd64 cmd/write-hotspot/*.go

customer-create:
	go build -o bin/customer-create cmd/customer-create/*.go
	GOOS=linux GOARCH=amd64 go build -o bin/customer-create-linux-amd64 cmd/customer-create/*.go