TARGET=pe
GOBIN=$(GOPATH)/bin

tools:
	@echo $(GOBIN)
	go install golang.org/x/tools/cmd/goimports@v0.8.0
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

install:
	mv ./pe /usr/local/bin/
	mkdir -p /usr/local/share/doc/pe
	cp ./doc/help.txt /usr/local/share/doc/pe/
	echo "Pine editor installed"
