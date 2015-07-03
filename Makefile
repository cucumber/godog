.PHONY: test deps

test:
	@echo "running all tests"
	@go fmt ./...
	@golint ./...
	go vet ./...
	go test
	go run cmd/godog/main.go -f progress

deps:
	@echo "updating all dependencies"
	go get -u github.com/cucumber/gherkin-go
	go get -u github.com/shiena/ansicolor

