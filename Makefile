.PHONY: help test vet fmt build run demo validate-demo serve-demo clean
help:
	@echo "Targets: fmt vet test build demo validate-demo serve-demo clean"
fmt:
	gofmt -w .
vet:
	go vet ./...
test:
	go test ./...
build:
	go build -o ./bin/nomos ./cmd/nomos
demo: build
	rm -rf ./tmp/demo-cosmos
	./bin/nomos cosmos init ./tmp/demo-cosmos --git
	./bin/nomos domain add identity.blumer.cloud --path ./tmp/demo-cosmos --owner "Identity Team"
	./bin/nomos service add user-account --domain identity.blumer.cloud --path ./tmp/demo-cosmos --owner "Identity Team"
validate-demo: build
	./bin/nomos validate --path ./tmp/demo-cosmos
serve-demo: build
	./bin/nomos serve --path ./tmp/demo-cosmos --listen 127.0.0.1:8080
clean:
	rm -rf ./bin ./tmp
