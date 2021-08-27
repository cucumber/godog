.PHONY: test gherkin bump cover

VERS := $(shell grep 'const Version' -m 1 godog.go | awk -F\" '{print $$2}')

test:
	@echo "running all tests"
	@go install ./...
	@go fmt ./...
	@golint github.com/cucumber/godog
	@golint github.com/cucumber/godog/cmd/godog
	go vet ./...
	go test -race
	godog -f progress -c 4

gherkin:
	@if [ -z "$(VERS)" ]; then echo "Provide gherkin version like: 'VERS=commit-hash'"; exit 1; fi
	@rm -rf gherkin
	@mkdir gherkin
	@curl -s -L https://github.com/cucumber/gherkin-go/tarball/$(VERS) | tar -C gherkin -zx --strip-components 1
	@rm -rf gherkin/{.travis.yml,.gitignore,*_test.go,gherkin-generate*,*.razor,*.jq,Makefile,CONTRIBUTING.md}

bump:
	@if [ -z "$(VERSION)" ]; then echo "Provide version like: 'VERSION=$(VERS) make bump'"; exit 1; fi
	@echo "bumping version from: $(VERS) to $(VERSION)"
	@sed -i.bak 's/$(VERS)/$(VERSION)/g' godog.go
	@sed -i.bak 's/$(VERS)/$(VERSION)/g' _examples/api/features/version.feature
	@find . -name '*.bak' | xargs rm

cover:
	go test -race -coverprofile=coverage.txt
	go tool cover -html=coverage.txt
	rm coverage.txt

ARTIFACT_DIR := _artifacts

# To upload artifacts for the current version;
# execute: make upload
#
# Check https://github.com/tcnksm/ghr for usage of ghr
upload: artifacts
	ghr -replace $(VERS) $(ARTIFACT_DIR)

# To build artifacts for the current version;
# execute: make artifacts
artifacts: 
	rm -rf $(ARTIFACT_DIR)
	mkdir $(ARTIFACT_DIR)

	$(call _build,darwin,amd64)
	$(call _build,linux,amd64)
	$(call _build,linux,arm64)


# To update packages from common monorepo;
# execute: make common
common:
	rm -rf ./_cucumber-common
	git clone https://github.com/cucumber/common.git --depth 1 ./_cucumber-common
	mkdir -p ./internal/messages
	mkdir -p ./internal/gherkin
	cp -f ./_cucumber-common/messages/go/*.go ./internal/messages/
	cp -f ./_cucumber-common/gherkin/go/*.go ./internal/gherkin/
	rm -rf ./_cucumber-common
	find . -type f -path '*/.go' -print0 | xargs -0 perl -i -pe "s|github.com/cucumber/common/messages/go/v\d+|github.com/cucumber/godog/internal/messages|g"
	find . -type f -path '*/.go' -print0 | xargs -0 perl -i -pe "s|github.com/cucumber/common/gherkin/go/v\d+|github.com/cucumber/godog/internal/gherkin|g"
	find ./internal/gherkin -type f -print0 | xargs -0 perl -i -pe "s|ExampleCompilePickles|ExamplePickles|g"
	git add ./internal/messages
	git add ./internal/gherkin

define _build
	mkdir $(ARTIFACT_DIR)/godog-$(VERS)-$1-$2
	env GOOS=$1 GOARCH=$2 go build -o $(ARTIFACT_DIR)/godog-$(VERS)-$1-$2/godog ./cmd/godog
	cp README.md $(ARTIFACT_DIR)/godog-$(VERS)-$1-$2/README.md
	cp LICENSE $(ARTIFACT_DIR)/godog-$(VERS)-$1-$2/LICENSE
	cd $(ARTIFACT_DIR) && tar -c --use-compress-program="pigz --fast" -f godog-$(VERS)-$1-$2.tar.gz godog-$(VERS)-$1-$2 && cd ..
	rm -rf $(ARTIFACT_DIR)/godog-$(VERS)-$1-$2
endef
