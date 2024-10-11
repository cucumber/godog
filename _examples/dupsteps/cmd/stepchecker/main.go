package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

//
// See accompanying README file(s);
// also https://github.com/cucumber/godog/pull/642
//

func main() {
	if len(os.Args) < 3 {
		log.Printf("Usage: main.go [go-file(s)] [feature-file(s)]\n")

		os.Exit(RC_USER)
	}

	// Structures into which to collect step patterns found in Go and feature files
	godogSteps := make(map[string]*StepMatch)
	featureSteps := make(map[string]*StepMatch)

	// collect input files (must have at least one of each kind (e.g., *.go, *.feature)
	for _, filePath := range os.Args[1:] {
		if strings.HasSuffix(filePath, ".go") {
			if err := collectGoSteps(filePath, godogSteps); err != nil {
				fmt.Printf("error collecting `go` steps: %s\n", err)

				os.Exit(RC_ISSUES)
			}
		}

		if strings.HasSuffix(filePath, ".feature") {
			if err := collectFeatureSteps(filePath, featureSteps); err != nil {
				fmt.Printf("error collecting `feature` steps: %s\n", err)

				os.Exit(RC_ISSUES)
			}
		}
	}

	if len(godogSteps) == 0 {
		log.Printf("no godog step definition(s) found")

		os.Exit(RC_USER)
	}

	if len(featureSteps) == 0 {
		log.Printf("no feature step invocation(s) found")

		os.Exit(RC_USER)
	}

	// Match steps between Go and feature files
	matchSteps(godogSteps, featureSteps)

	var issuesFound int

	// Report on unexpected (i.e., lack of, duplicate or ambiguous) mapping from feature steps to go steps
	fmt.Printf("Found %d feature step(s):\n", len(featureSteps))

	var fsIdx int

	for text, step := range featureSteps {
		fsIdx++

		fmt.Printf("%d. %q\n", fsIdx, text)

		if len(step.matchedWith) != 1 {
			issuesFound++

			fmt.Printf("  - %d matching godog step(s) found:\n", len(step.matchedWith))

			for _, match := range step.source {
				fmt.Printf("    from: %s\n", match)
			}

			for _, match := range step.matchedWith {
				fmt.Printf("      to: %s\n", match)
			}
		}
	}

	fmt.Println()

	// Report on lack of mapping from go steps to feature steps
	fmt.Printf("Found %d godog step(s):\n", len(godogSteps))

	var gdsIdx int

	for text, step := range godogSteps {
		gdsIdx++

		fmt.Printf("%d. %q\n", gdsIdx, text)

		if len(step.matchedWith) == 0 {
			issuesFound++

			fmt.Printf("  - No matching feature step(s) found:\n")

			for _, match := range step.source {
				fmt.Printf("    from: %s\n", match)
			}
		}
	}

	fmt.Println()

	if issuesFound != 0 {
		log.Printf("%d issue(s) found\n", issuesFound)
		os.Exit(RC_ISSUES)
	}
}

func collectGoSteps(goFilePath string, steps map[string]*StepMatch) error {
	fset := token.NewFileSet()

	node, errParse := parser.ParseFile(fset, goFilePath, nil, parser.ParseComments)
	if errParse != nil {
		return fmt.Errorf("error parsing Go file %s: %w", goFilePath, errParse)
	}

	stepDefPattern := regexp.MustCompile(`^godog.ScenarioContext.(Given|When|Then|And|Step)$`)

	var errInspect error

	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		methodID := extractMethodID(call)
		if methodID == "" {
			return true
		}

		if !stepDefPattern.MatchString(methodID) {
			return true
		}

		if len(call.Args) == 0 {
			log.Printf("WARNING: ignoring call to step function with no arguments: %s\n", methodID)

			return true
		}

		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			log.Printf("WARNING: ignoring unexpected step function invocation at %s\n", fset.Position(call.Pos()))

			return true
		}

		pattern, errQ := strconv.Unquote(lit.Value)

		if errQ != nil {
			errInspect = errQ
			return false
		}

		sm, found := steps[pattern]
		if !found {
			sm = &StepMatch{}
			steps[pattern] = sm
		}

		sm.source = append(sm.source, sourceRef(fset.Position(lit.ValuePos).String()))

		return true
	})

	if errInspect != nil {
		return fmt.Errorf("error encountered while inspecting %q: %w", goFilePath, errInspect)
	}

	return nil
}

func collectFeatureSteps(featureFilePath string, steps map[string]*StepMatch) error {
	file, errOpen := os.Open(featureFilePath)
	if errOpen != nil {
		return errOpen
	}

	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	sp := regexp.MustCompile(`^\s*(Given|When|Then|And) (.+)\s*$`)

	for lineNo := 1; scanner.Scan(); lineNo++ {

		line := scanner.Text()
		matches := sp.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		if len(matches) != 3 {
			return fmt.Errorf("unexpected number of matches at %s:%d: %d for %q\n", featureFilePath, lineNo, len(matches), line)
		}

		stepText := matches[2]

		sm, found := steps[stepText]
		if !found {
			sm = &StepMatch{}
			steps[stepText] = sm
		}

		sm.source = append(sm.source, sourceRef(fmt.Sprintf("%s:%d", featureFilePath, lineNo)))
	}

	return nil
}

func matchSteps(godogSteps, featureSteps map[string]*StepMatch) {
	// for each step definition found in go
	for pattern, godogStep := range godogSteps {
		matcher, errComp := regexp.Compile(pattern)
		if errComp != nil {
			log.Printf("error compiling regex for pattern '%s': %v\n", pattern, errComp)

			continue
		}

		// record matches between feature steps and go steps
		for featureText, featureStep := range featureSteps {
			if matcher.MatchString(featureText) {
				featureStep.matchedWith = append(featureStep.matchedWith, godogStep.source...)
				godogStep.matchedWith = append(godogStep.matchedWith, featureStep.source...)
			}
		}
	}
}

func extractMethodID(call *ast.CallExpr) string {
	// Ensure the function is a method call
	fnSelExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	// Ensure the method is being called on an identifier
	lhsID, ok := fnSelExpr.X.(*ast.Ident)
	if !ok || lhsID.Obj == nil {
		return ""
	}

	// Ensure the identifier represents a field declaration
	lhsField, ok := lhsID.Obj.Decl.(*ast.Field)
	if !ok {
		return ""
	}

	// Ensure the field type is a pointer to another type
	lhsStarExpr, ok := lhsField.Type.(*ast.StarExpr)
	if !ok {
		return ""
	}

	// Ensure the pointer type is a package or struct
	lhsSelExpr, ok := lhsStarExpr.X.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	// Ensure the receiver type is an identifier (e.g., the package or struct name)
	lhsLhsID, ok := lhsSelExpr.X.(*ast.Ident)
	if !ok {
		return ""
	}

	// return a method call identifier sufficient to identify those of interest
	return fmt.Sprintf("%s.%s.%s", lhsLhsID.Name, lhsSelExpr.Sel.Name, fnSelExpr.Sel.Name)
}

type sourceRef string

type StepMatch struct {
	source      []sourceRef // location of the source(s) for this step
	matchedWith []sourceRef // location of match(es) to this step
}

const RC_ISSUES = 1
const RC_USER = 2
