package cpu

import (
	"fmt"

	"github.com/sshaman1101/uvm/asm"
)

const (
	ROMSize       = 1 << 16
	RAMSize       = 512
	RegisterCount = 8
	StackDepth    = 32
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

	// op codes for this instance
	opCodes *asm.OpCodes
}

func NewCPU(codes *asm.OpCodes) *CPU {
	return &CPU{
		ROM:       [ROMSize]uint8{},
		RAM:       [RAMSize]uint8{},
		registers: [8]uint8{},
		stack:     newStack(StackDepth),
		flags:     &flags{},
		pc:        0,
		opCodes:   codes,
	}
}

func (cpu *CPU) reset() {
	cpu = NewCPU(cpu.opCodes)
}

func (cpu *CPU) Run() {
	for {
		// load next value from mem,
		// must be an instruction
		v := cpu.ROM[cpu.pc]

		// decode instruction
		// note: panics on invalid input
		nextInstruction := cpu.decodeInstruction(v)

		// debug
		fmt.Printf("instruction %s at %d\n", nextInstruction.name, cpu.pc)

		// note: just for testing purposes now re're using
		//  PC as MAR/MDR to load operands, I'll fix that later (probably).
		// add offset to PC, next instruction must be at
		// PC+(operand count), thus we assume each operand size is one uint8
		// cpu.pc += nextInstruction.operandCount

		// load operands
		for i := 0; i < nextInstruction.operandCount; i++ {
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
			fmt.Printf("  r%d = %02x", i, cpu.registers[i])
		}
		fmt.Println()
	}
}
