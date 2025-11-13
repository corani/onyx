package obsidian

// Diary manages the `## One Line` section of a daily note.
type Diary struct {
	Doc *Document
	Sec *Section
}

func NewDiary(doc *Document) (*Diary, error) {
	sec, err := doc.Section("## One Line")
	if err != nil {
		return nil, err
	}

	return &Diary{
		Doc: doc,
		Sec: sec,
	}, nil
}

// Get returns the diary section body (without the header).
func (d *Diary) Get() string {
	return d.Sec.Body()
}

// Set replaces the diary body while preserving section newline invariants.
// Whitespace-only bodies are treated as empty; non-empty bodies are preserved verbatim.
func (d *Diary) Set(body string) error {
	d.Sec.SetBody(body)

	return d.Doc.Save(false)
}
