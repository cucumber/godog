GOOD_FEATURE_FILES = $(shell find ../testdata/good -name "*.feature")
BAD_FEATURE_FILES  = $(shell find ../testdata/bad -name "*.feature")

TOKENS   = $(patsubst ../testdata/%.feature,acceptance/testdata/%.feature.tokens,$(GOOD_FEATURE_FILES))
ASTS     = $(patsubst ../testdata/%.feature,acceptance/testdata/%.feature.ast.json,$(GOOD_FEATURE_FILES))
ERRORS   = $(patsubst ../testdata/%.feature,acceptance/testdata/%.feature.errors,$(BAD_FEATURE_FILES))

GO_SOURCE_FILES = $(shell find . -name "*.go") parser.go dialects_builtin.go

export GOPATH = $(realpath ./)

all: .compared

test: $(TOKENS) $(ASTS) $(ERRORS)
.PHONY: test

.compared: .built $(TOKENS) $(ASTS) $(ERRORS)
	touch $@

.built: show-version-info $(GO_SOURCE_FILES) bin/gherkin-generate-tokens bin/gherkin-generate-ast LICENSE
	touch $@

show-version-info:
	go version

bin/gherkin-generate-tokens: $(GO_SOURCE_FILES)
	go build -o $@ ./gherkin-generate-tokens

bin/gherkin-generate-ast: $(GO_SOURCE_FILES)
	go build -o $@ ./gherkin-generate-ast

acceptance/testdata/%.feature.tokens: ../testdata/%.feature ../testdata/%.feature.tokens bin/gherkin-generate-tokens
	mkdir -p `dirname $@`
	bin/gherkin-generate-tokens $< > $@
	diff --unified $<.tokens $@
.DELETE_ON_ERROR: acceptance/testdata/%.feature.tokens

acceptance/testdata/%.feature.ast.json: ../testdata/%.feature ../testdata/%.feature.ast.json bin/gherkin-generate-ast
	mkdir -p `dirname $@`
	bin/gherkin-generate-ast $< | jq --sort-keys "." > $@
	diff --unified $<.ast.json $@
.DELETE_ON_ERROR: acceptance/testdata/%.feature.ast.json

acceptance/testdata/%.feature.errors: ../testdata/%.feature ../testdata/%.feature.errors bin/gherkin-generate-ast
	mkdir -p `dirname $@`
	! bin/gherkin-generate-ast $< 2> $@
	diff --unified $<.errors $@
.DELETE_ON_ERROR: acceptance/testdata/%.feature.errors

parser.go: ../gherkin.berp parser.go.razor ../bin/berp.exe
	mono ../bin/berp.exe -g ../gherkin.berp -t parser.go.razor -o $@
	# Remove BOM
	tail -c +4 $@ > $@.nobom
	mv $@.nobom $@

dialects_builtin.go: ../gherkin-languages.json dialects_builtin.go.jq
	cat $< | jq -f dialects_builtin.go.jq -r -c > $@

LICENSE: ../LICENSE
	cp $< $@

clean:
	rm -rf .compared .built acceptance bin/ parser.go dialects_builtin.go
.PHONY: clean show-version-info

update-gherkin-languages: dialects_builtin.go
.PHONY: update-gherkin-languages
