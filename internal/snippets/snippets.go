package snippets

import "github.com/cucumber/godog/internal/storage"

// Func defines an interface for functions that render snippets
type Func func(*storage.Storage) string

var registry map[string]Func

func init() {
	registry = make(map[string]Func)
}

func register(name string, f Func) {
	registry[name] = f
}

// Find finds a registered snippet function and returns it
//
// If the snippet function is not found, the original function is returned
func Find(name string) Func {
	f, ok := registry[name]
	if ok {
		return f
	}
	return StepFunc
}
