.PHONY: all test zip test-config clean

PKG=github.com/euank/buildkite-kettle

GO_FILES=$(shell find . -name "*.go")
 
BUILDKITE_CONFIG=$(shell cat config.json)

all: ./bin/handler

test:
	go test -race $(PKG)/...

bin/handler: $(GO_FILES)
	mkdir -p bin
	go build -o ./bin/handler main.go

zip: bin/handler test-config deploy/func.zip
	@echo Please upload 'deploy/func.zip' to lambda

test-config:
	TEST_CONFIG=true ./bin/handler

deploy/func.zip: bin/handler config.json
	mkdir -p deploy
	rm -f deploy/func.zip
	zip deploy/func.zip bin/handler config.json
	

clean:
	rm -f ./bin/handler
	rm -f deploy/func.zip
