package key

import (
	"bytes"
	"fmt"
)

// Subspace ...
type Subspace interface {
	Sub(el ...TupleElement) Subspace
	Bytes() []byte
	Pack(t Tuple) Key
}

type subspace struct {
	rawPrefix []byte
}

func (s subspace) Sub(el ...TupleElement) Subspace {
	return subspace{concat(s.Bytes(), Tuple(el).Pack()...)}
}

func (s subspace) Bytes() []byte {
	return s.rawPrefix
}

func (s subspace) Pack(t Tuple) Key {
	return Key(concat(s.rawPrefix, t.Pack()...))
}

type pathspace struct {
	path []byte
}

func (p pathspace) Sub(el ...TupleElement) Subspace {
	b := concat(p.Bytes(), []byte("/")...)
	return pathspace{concat(b, Tuple(el).Pack()...)}
}

func (p pathspace) Bytes() []byte {
	return p.path
}

func (p pathspace) Pack(t Tuple) Key {
	b := concat(p.path, []byte("/")...)
	return Key(concat(b, t.Pack()...))
}

// FromBytes returns a new Subspace from the provided bytes.
func FromBytes(b []byte) Subspace {
	s := make([]byte, len(b))
	copy(s, b)
	return subspace{s}
}

func NewPath() Subspace {
	return pathspace{[]byte{}}
}

func concat(a []byte, b ...byte) []byte {
	r := make([]byte, len(a)+len(b))
	copy(r, a)
	copy(r[len(a):], b)
	return r
}

// Printable ...
func Printable(d []byte) string {
	buf := new(bytes.Buffer)
	for _, b := range d {
		if b >= 32 && b < 127 && b != '\\' {
			buf.WriteByte(b)
			continue
		}
		if b == '\\' {
			buf.WriteString("\\\\")
			continue
		}
		buf.WriteString(fmt.Sprintf("\\x%02x", b))
	}
	return buf.String()
}
