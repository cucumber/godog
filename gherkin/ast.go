package gherkin

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Node struct {
	Location *Location `json:"location,omitempty"`
	Type     string    `json:"type"`
}

type Feature struct {
	Node
	Tags                []*Tag        `json:"tags"`
	Language            string        `json:"language,omitempty"`
	Keyword             string        `json:"keyword"`
	Name                string        `json:"name"`
	Description         string        `json:"description,omitempty"`
	Background          *Background   `json:"background,omitempty"`
	ScenarioDefinitions []interface{} `json:"scenarioDefinitions"`
	Comments            []*Comment    `json:"comments"`
}

type Comment struct {
	Node
	Text string `json:"text"`
}

type Tag struct {
	Node
	Name string `json:"name"`
}

type Background struct {
	ScenarioDefinition
}

type Scenario struct {
	ScenarioDefinition
	Tags []*Tag `json:"tags"`
}

type ScenarioOutline struct {
	ScenarioDefinition
	Tags     []*Tag      `json:"tags"`
	Examples []*Examples `json:"examples,omitempty"`
}

type Examples struct {
	Node
	Tags        []*Tag      `json:"tags"`
	Keyword     string      `json:"keyword"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	TableHeader *TableRow   `json:"tableHeader"`
	TableBody   []*TableRow `json:"tableBody"`
}

type TableRow struct {
	Node
	Cells []*TableCell `json:"cells"`
}

type TableCell struct {
	Node
	Value string `json:"value"`
}

type ScenarioDefinition struct {
	Node
	Keyword     string  `json:"keyword"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Steps       []*Step `json:"steps"`
}

type Step struct {
	Node
	Keyword    string      `json:"keyword"`
	Text       string      `json:"text"`
	Argument   interface{} `json:"argument,omitempty"`
	Embeddings `json:"embeddings,omitempty"`
}

type Embeddings struct {
	EmbeddedContent []*Embedding
}

func (e *Embeddings) AddEmbedding(mimeType string, data string) {
	if e.EmbeddedContent == nil {
		e.EmbeddedContent = make([]*Embedding, 0)
	}
	e.EmbeddedContent = append(e.EmbeddedContent, &Embedding{mimeType, data})
}

type Embedding struct {
	MimeType string `json:"mime_type"`
	Data     string `json:"data"`
}

type DocString struct {
	Node
	ContentType string `json:"contentType,omitempty"`
	Content     string `json:"content"`
	Delimitter  string `json:"-"`
}

type DataTable struct {
	Node
	Rows []*TableRow `json:"rows"`
}
