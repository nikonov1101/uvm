package cpu

import (
	"fmt"
)

const (
	ROMSize       = 512
	RAMSize       = 512
	RegisterCount = 8
)

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
	ROM [ROMSize]uint8
	RAM [RAMSize]uint8

	registers [8]uint8
	stack     *stack
	flags     *flags

	// program counter
	pc uint16
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
		//  PC as MAR/MDR to load operands, I'll fix that later (probably).
		// add offset to PC, next instruction must be at
		// PC+(operand count), thus we assume each operand size is one uint8
		// cpu.pc += nextInstruction.operandCount

		// load operands
		for i := uint16(0); i < nextInstruction.operandCount; i++ {
			// calculate next mem address
			cpu.pc++
			// fetch memory
			// todo: do it via something like MAR/MDR, as real hardware do
			//  or just emulate it, at least it will looks hacky.
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
		for i := 0; i < RegisterCount; i++ {
			fmt.Printf("  r%d = %v", i, cpu.registers[i])
		}
		fmt.Println()
	}
}
