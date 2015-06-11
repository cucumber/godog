package godog

import "regexp"

var stepHandlers map[*regexp.Regexp]StepHandler

// Objects implementing the StepHandler interface can be
// registered as step definitions in godog
//
// HandleStep method receives all arguments which
// will be matched according to the regular expression
// which is passed with a step registration.
// The error in return - represents a reason of failure.
//
// Returning signals that the step has finished
// and that the feature runner can move on to the next
// step.
type StepHandler interface {
	HandleStep(args ...interface{}) error
}

// StepHandlerFunc type is an adapter to allow the use of
// ordinary functions as Step handlers.  If f is a function
// with the appropriate signature, StepHandlerFunc(f) is a
// StepHandler object that calls f.
type StepHandlerFunc func(...interface{}) error

// HandleStep calls f(step_arguments...).
func (f StepHandlerFunc) HandleStep(args ...interface{}) error {
	return f(args...)
}

// Step registers a StepHandler which will be triggered
// if regular expression will match a step from a feature file.
//
// If none of the StepHandlers are matched, then a pending
// step error will be raised.
func Step(exp *regexp.Regexp, h StepHandler) {
	stepHandlers[exp] = h
}

func init() {
	stepHandlers = make(map[*regexp.Regexp]StepHandler)
}
