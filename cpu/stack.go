package cpu

// just a simple fixed-depth stack
type stack struct {
	data    []uint8
	pointer int
}

func (s *stack) push(v uint8) {
	if s.pointer >= len(s.data) {
		panic("stack overflow")
	}
	s.data[s.pointer] = v
	s.pointer++
}

func (s *stack) pop() uint8 {
	if s.pointer == 0 {
		panic("stack is empty")
	}

	s.pointer--
	return s.data[s.pointer]
}

func newStack(depth int) *stack {
	return &stack{
		pointer: 0,
		data:    make([]uint8, depth),
	}
}
