package snippets

import (
	"bytes"
	"fmt"
	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/storage"
	"sort"
	"strings"
	"unicode"
)

func init() {
	register("step_func", StepFunction)
}

func StepFunction(s *storage.Storage) string {
	undefinedStepResults := s.MustGetPickleStepResultsByStatus(models.Undefined)
	if len(undefinedStepResults) == 0 {
		return ""
	}

	var index int
	var snips []undefinedSnippet
	// build snippets
	for _, u := range undefinedStepResults {
		pickleStep := s.MustGetPickleStep(u.PickleStepID)

		steps := []string{pickleStep.Text}
		arg := pickleStep.Argument
		if u.Def != nil {
			steps = u.Def.Undefined
			arg = nil
		}
		for _, step := range steps {
			expr := snippetExprCleanup.ReplaceAllString(step, "\\$1")
			expr = snippetNumbers.ReplaceAllString(expr, "(\\d+)")
			expr = snippetExprQuoted.ReplaceAllString(expr, "$1\"([^\"]*)\"$2")
			expr = "^" + strings.TrimSpace(expr) + "$"

			name := snippetNumbers.ReplaceAllString(step, " ")
			name = snippetExprQuoted.ReplaceAllString(name, " ")
			name = strings.TrimSpace(snippetMethodName.ReplaceAllString(name, ""))
			var words []string
			for i, w := range strings.Split(name, " ") {
				switch {
				case i != 0:
					w = strings.Title(w)
				case len(w) > 0:
					w = string(unicode.ToLower(rune(w[0]))) + w[1:]
				}
				words = append(words, w)
			}
			name = strings.Join(words, "")
			if len(name) == 0 {
				index++
				name = fmt.Sprintf("StepDefinitioninition%d", index)
			}

			var found bool
			for _, snip := range snips {
				if snip.Expr == expr {
					found = true
					break
				}
			}
			if !found {
				snips = append(snips, undefinedSnippet{Method: name, Expr: expr, argument: arg})
			}
		}
	}

	sort.Sort(snippetSortByMethod(snips))

	var buf bytes.Buffer
	if err := undefinedSnippetsTpl.Execute(&buf, snips); err != nil {
		panic(err)
	}
	// there may be trailing spaces
	return strings.Replace(buf.String(), " \n", "\n", -1)
}
