TARGET=pe
GOBIN=$(GOPATH)/bin

tools:
	@echo $(GOBIN)
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.2

.PHONY: fmt
fmt:
	go fmt ./...
.PHONY: lint
lint:
	$(GOBIN)/golangci-lint run --fix
.PHONY: import
import:
	$(GOBIN)/goimports -l -w .

clean:
	rm $(TARGET)

.PHONY: build
build:
	@echo $(GOFILES)
	go build -o $(TARGET) ./cmd/pine/

.PHONY: run
run:
	go run cmd/pine/main.go

.PHONY: all
all: fmt lint import build
