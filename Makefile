build:
	@echo $(GOFILES)
	go build -o pine cmd/pine/main.go
run:
	go run cmd/pine/main.go
