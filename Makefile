.PHONY: test

# runs all necessary tests
test:
	@sh -c 'if [ ! -z "$(go fmt ./...)" ]; then exit 1; fi'
	golint ./...
	go vet ./...
	go test ./...
	go run cmd/godog/main.go -f progress

# updates dependencies
deps:
	go get -u github.com/cucumber/gherkin-go
	go get -u golang.org/x/tools/imports
	go get -u github.com/shiena/ansicolor
