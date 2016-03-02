.PHONY: test deps

test:
	@echo "running all tests"
	@go fmt ./...
	@golint ./...
	go vet ./...
	go test
	go run cmd/godog/main.go -f progress -c 4

deps:
	@echo "updating all dependencies"
	go get -u gopkg.in/cucumber/gherkin-go.v3
	go get -u github.com/shiena/ansicolor

