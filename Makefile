GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=kaboo

all: test build
build:
	$(GOBUILD) -o ./bin/$(BINARY_NAME) -v ./cmd/kaboo-server-go
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/kaboo-server-go
	./bin/$(BINARY_NAME)