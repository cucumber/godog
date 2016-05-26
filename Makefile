.PHONY: test gherkin

test:
	@echo "running all tests"
	@go install ./...
	@go fmt ./...
	@golint github.com/DATA-DOG/godog
	@golint github.com/DATA-DOG/godog/cmd/godog
	go vet ./...
	go test
	godog -f progress -c 4

gherkin:
	@if [ -z "$(VERS)" ]; then echo "Provide gherkin version like: 'VERS=commit-hash'"; exit 1; fi
	@rm -rf gherkin
	@mkdir gherkin
	@curl -s -L https://github.com/cucumber/gherkin-go/tarball/$(VERS) | tar -C gherkin -zx --strip-components 1
	@rm -rf gherkin/{.travis.yml,.gitignore,*_test.go,gherkin-generate*,*.razor,*.jq,Makefile,CONTRIBUTING.md}
