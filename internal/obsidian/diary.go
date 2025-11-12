package obsidian

// Diary manages the `## One Line` section of a daily note using a shared Document.
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

// Get returns the body of the diary section (without the header).
func (d *Diary) Get() string {
	return d.Sec.Body()
}

// Set replaces the entire diary section with the provided body. The section is written as:
//
//	## One Line
//	<blank line>
//	<body>
//	<blank line> (only if body is non-empty)
//
// If body is empty (or whitespace), only a single blank line after the header is written.
// The user's body is preserved verbatim (no wrapping or trimming).
func (d *Diary) Set(body string) error {
	d.Sec.SetBody(body)

	return d.Doc.Save(false)
}
