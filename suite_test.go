package godog

// needed in order to use godog cli
func GodogContext(s *Suite) {
	SuiteContext(s)
}
