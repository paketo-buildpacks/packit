package sbom

import "io"

type Entries struct {
	content SBOM
	formats map[Format]string
}

func NewEntries(content SBOM) Entries {
	return Entries{
		content: content,
		formats: make(map[Format]string),
	}
}

func (e Entries) Format() (map[string]io.Reader, error) {
	result := make(map[string]io.Reader)
	for format := range e.formats {
		result[format.Extension()] = e.content.Format(format)
	}
	return result, nil
}

func (e Entries) IsEmpty() bool {
	return e.content.IsEmpty()
}

func (e Entries) AddFormat(format Format) {
	e.formats[format] = ""
}

func (e Entries) GetContent(format Format) io.Reader {
	return e.content.Format(format)
}
