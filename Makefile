build:
	@echo $(GOFILES)
	go build -o ste cmd/ste/main.go
run:
	go run cmd/ste/main.go
