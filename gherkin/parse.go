package gherkin

type Tag string

type Scenario struct {
	Steps []*Step
	Tags  []Tag
	Line  string
}

type Background struct {
	Steps []*Step
	Line  string
}

type StepType string

const (
	Given StepType = "Given"
	When  StepType = "When"
	Then  StepType = "Then"
)

type Step struct {
	Line string
	Text string
	Type StepType
}

type Feature struct {
	Tags        []Tag
	Description string
	Line        string
	Title       string
	Filename    string
	Background  *Background
	Scenarios   []*Scenario
}

// func Parse(r io.Reader) (*Feature, error) {
// 	in := bufio.NewReader(r)
// 	for line, err := in.ReadString(byte('\n')); err != nil; line, err = in.ReadString(byte('\n')) {
// 		ln := strings.TrimFunc(string(line), unicode.IsSpace)
// 	}
// 	return nil, nil
// }
