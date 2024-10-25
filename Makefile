
test-ddl:
	GOOS=linux GOARCH=amd64 go build -o bin/test-ddl cmd/test-ddl/*.go

write-hotspot:
	GOOS=linux GOARCH=amd64 go build -o bin/write-hotspot cmd/write-hotspot/*.go

customer-create:
	GOOS=linux GOARCH=amd64 go build -o bin/customer-create cmd/customer-create/*.go