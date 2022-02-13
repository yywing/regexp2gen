package regexp2gen

import (
	"bytes"
)

type Buffer struct {
	*bytes.Buffer
	buffers []*bytes.Buffer
	marks   map[int]*bytes.Buffer
}

func NewBuffer() *Buffer {
	return &Buffer{
		Buffer:  &bytes.Buffer{},
		buffers: []*bytes.Buffer{},
		marks:   make(map[int]*bytes.Buffer),
	}
}

func (b *Buffer) WriteAll(p []byte) (n int, err error) {
	cnt := 0
	for {
		i := cnt - 1
		if i < 0 {
			i = 0
		}
		n, err := b.Write(p[i:])
		cnt += n
		if err != nil {
			return cnt, err
		}
		if cnt == len(p) {
			return cnt, nil
		}
	}
}

// push
func (b *Buffer) Setmark() {
	b.buffers = append(b.buffers, b.Buffer)
	b.Buffer = &bytes.Buffer{}
}

// pop
func (b *Buffer) Backmark(capture bool, index int) error {
	outer := b.Buffer
	l := len(b.buffers)
	b.Buffer = b.buffers[l-1]
	b.buffers = b.buffers[:l-1]

	_, err := b.WriteAll(outer.Bytes())
	if err != nil {
		return err
	}

	if capture {
		b.marks[index] = outer
	}
	return nil
}

func (b *Buffer) Getmark(index int) (*bytes.Buffer, bool) {
	d, ok := b.marks[index]
	return d, ok
}
