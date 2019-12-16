package main

import (
	"fmt"
)

const (
	romSize       = 512
	ramSize       = 512
	registerCount = 8

	_ = iota
	operandValue
	operandRegister
	operandAddress
)

type operand struct {
	opType int
	value  uint8
}

// checkOperand checks that given value can be
// an operand of the expected type.
// returns human readable name for debug
func checkOperand(v uint8, typ int) string {
	switch typ {
	case operandValue:
		// value can be any value, nothing to do here
		return fmt.Sprintf("#%d", v)
	case operandRegister:
		if v >= registerCount {
			panic("invalid register operand")
		}
		return fmt.Sprintf("r%d", v)
	case operandAddress:
		if int(v) >= romSize {
			panic("mem address operand is out of memory")
		}
		return fmt.Sprintf("$%d", v)
	default:
		panic("operand type must be defined")
	}
}

type instruction struct {
	// just a name, like MOV, XOR, JUMP
	name string
	// what to do
	opCode uint8
	// how many operands we need to fetch from memory
	operandCount uint16
	// on which data we need to preform operation
	operands []operand
}

// execute the instruction
// can touch:
//   * registers
//   * flag register
//   * program counter
// note that in must increase PC by one
//   if it's regular instruction (not JUMP)
func (in *instruction) execute(cpu *CPU) {
	switch in.opCode {
	case asmNOP:
		// just do nothing

	case asmJUMP:
		// go to address, DO NOT increment PC by one
		cpu.pc = uint16(in.operands[0].value)
		return

	case asmADDRegReg:
		// todo: handle carry, overflow and zero
		// store result in the first operand register
		r0 := in.operands[0].value
		r1 := in.operands[1].value
		cpu.registers[r0] += cpu.registers[r1]

	case asmMOVRegVal:
		cpu.registers[in.operands[0].value] = in.operands[1].value

	case asmHALT:
		cpu.flags.halt = true
		return

	case asmPUSH:
		regVal := cpu.registers[in.operands[0].value]
		cpu.stack.push(regVal)
	case asmPOP:
		cpu.registers[in.operands[0].value] = cpu.stack.pop()

	default:
		panic(fmt.Sprintf("dunno how to execute instruction %2x", in.opCode))
	}

	// go to next instruction
	cpu.pc++
}

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

type flags struct {
	overflow bool
	zero     bool
	carry    bool
	halt     bool
}

func (f *flags) String() string {
	return fmt.Sprintf("O: %v | Z: %v | C: %v | H: %v", f.overflow, f.zero, f.carry, f.halt)
}

type CPU struct {
	ROM [romSize]uint8
	RAM [ramSize]uint8

	registers [8]uint8
	stack     *stack
	flags     *flags

	// program counter
	pc uint16
	// memory address register
	mar uint16
}

func NewCPU() *CPU {
	return &CPU{
		ROM:       [512]uint8{},
		RAM:       [512]uint8{},
		registers: [8]uint8{},
		stack:     newStack(32),
		flags:     &flags{},
		pc:        0,
	}
}

func (cpu *CPU) reset() {
	cpu = NewCPU()
}

func (cpu *CPU) Run() {
	for {
		// load next value from mem,
		// must be an instruction
		v := cpu.ROM[cpu.pc]

		// decode instruction
		// note: panics on invalid input
		nextInstruction := decodeInstruction(v)

		// debug
		fmt.Printf("instruction %s at %d\n", nextInstruction.name, cpu.pc)

		// note: just for testing purposes now re're using
		//  PC as MAR to load operands, I'll fix that later (probably).
		// add offset to PC, next instruction must be at
		// PC+(operand count), thus we assume each operand size is one uint8
		// cpu.pc += nextInstruction.operandCount

		// load operands
		for i := uint16(0); i < nextInstruction.operandCount; i++ {
			cpu.pc++
			given := cpu.ROM[cpu.pc]
			expected := nextInstruction.operands[i]

			// sanity check
			opName := checkOperand(given, expected.opType)

			fmt.Printf("  operand %s loaded\n", opName)

			// store withing instruction
			nextInstruction.operands[i].value = given
		}

		nextInstruction.execute(cpu)

		if cpu.flags.halt {
			fmt.Printf("CPU were HALTed at %2x\n", cpu.pc)
			return
		}

		// dump CPU state
		fmt.Printf("PC = %d\n", cpu.pc)
		fmt.Printf("flags = %v\n", cpu.flags)
		fmt.Printf("registers:\n")
		for i := 0; i < registerCount; i++ {
			fmt.Printf("  r%d = %v", i, cpu.registers[i])
		}
		fmt.Println()
	}
}

const (
	asmNOP  = 0x00
	asmJUMP = 0x01
	asmPUSH = 0x02
	asmPOP  = 0x03
	asmHALT = 0x09

	asmADDRegReg = 0x10
	asmADDRegVal = 0x11

	asmMOVRegReg  = 0x20
	asmMOVRegVal  = 0x21
	asmMOVRegAddr = 0x22
)

func decodeInstruction(v uint8) instruction {
	var m = map[uint8]instruction{
		// do nothing
		asmNOP: {name: "NOP", operandCount: 0},
		// load new address to PC
		asmJUMP: {name: "JUMP", operandCount: 1, operands: []operand{{opType: operandAddress}}},
		// PUSH r1
		asmPUSH: {name: "PUSH", operandCount: 1, operands: []operand{{opType: operandRegister}}},
		// POP r1
		asmPOP: {name: "POP", operandCount: 1, operands: []operand{{opType: operandRegister}}},
		// stop all the things
		asmHALT: {name: "HALT", operandCount: 0},

		// ADD r1 + r2
		asmADDRegReg: {name: "ADD", operandCount: 2, operands: []operand{{opType: operandRegister}, {opType: operandRegister}}},
		// ADD reg + #val
		asmADDRegVal: {name: "ADD", operandCount: 2, operands: []operand{{opType: operandRegister}, {opType: operandValue}}},
		// MOV r1 <- r2
		asmMOVRegReg: {name: "MOV", operandCount: 2, operands: []operand{{opType: operandRegister}, {opType: operandRegister}}},
		// MOV r1 <- #val
		asmMOVRegVal: {name: "MOV", operandCount: 2, operands: []operand{{opType: operandRegister}, {opType: operandValue}}},
		// MOV r1 <- $addr
		asmMOVRegAddr: {name: "MOV", operandCount: 2, operands: []operand{{opType: operandRegister}, {opType: operandAddress}}},
	}

	ins, ok := m[v]
	if !ok {
		panic(fmt.Sprintf("invalid instruiction %2x", v))
	}

	ins.opCode = v
	return ins
}

var prog = [romSize]uint8{
	0x21, 0x00, 0x03, // mov r0, #3
	0x21, 0x01, 0x02, // mov r1, #2
	0x10, 0x00, 0x01, // add r0, r1 (store result in r1)
	0x00, 0x00, // NOP, NOP
	0x02, 0x00, // PUSH r1
	0x00, 0x00, // NOP, NOP
	0x03, 0x05, // POP r5
	0x09, // HALT
}

func main() {
	uCPU := NewCPU()
	uCPU.ROM = prog
	uCPU.Run()
}
