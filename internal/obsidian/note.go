package obsidian

type Note struct {
	Doc *Document
	Sec *Section
}

func NewNote(doc *Document) (*Note, error) {
	sec, err := doc.Section("## Notes")
	if err != nil {
		return nil, err
	}

	return &Note{
		Doc: doc,
		Sec: sec,
	}, nil
}

func (n *Note) List() string {
	body := n.Sec.Body()
	if body != "" {
		body = "\n" + body + "\n"
	}

	return n.Sec.Header + "\n" + body
}

func (n *Note) Create(text string) error {
	n.Sec.AppendLine(text)

	return n.Doc.Save(false)
}
