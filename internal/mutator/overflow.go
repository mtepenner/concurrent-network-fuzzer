package mutator

import "bytes"

type BufferOverflowMutator struct {
	size int
}

func NewBufferOverflowMutator(size int) *BufferOverflowMutator {
	return &BufferOverflowMutator{size: size}
}

func (m *BufferOverflowMutator) Generate() []byte {
	return bytes.Repeat([]byte("A"), m.size)
}

func (m *BufferOverflowMutator) Name() string {
	return "BufferOverflow"
}
