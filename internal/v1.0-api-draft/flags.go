package godog

// BindCommandLineFlags binds godog flags to given flag set prefixed
// by given prefix, without overriding usage
func BindCommandLineFlags(prefix string, opts *Options)
