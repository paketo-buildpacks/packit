package scribe

import (
	"io"

	"github.com/cheggaaa/pb/v3"
)

type Bar struct {
	w io.Writer
	p *pb.ProgressBar
}

func NewBar(w io.Writer) *Bar {
	return &Bar{
		w: w,
		p: pb.New(100),
	}
}

func (b *Bar) Start() {
	b.p.Start()
	b.p.SetWriter(b.w)
	b.p.SetWidth(100)
	b.p.SetTemplateString(`{{bar . "[" "-" ">" " " "]"}} {{percent .}} {{etime .}}`)
}

func (b *Bar) Increment() {
	b.p.Increment()
}

func (b *Bar) Finish() {
	b.p.Finish()
}
