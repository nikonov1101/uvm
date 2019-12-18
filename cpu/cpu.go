package cpu

import (
	"fmt"

	"github.com/sshaman1101/uvm/asm"
	"github.com/sshaman1101/uvm/defines"
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
	ROM [defines.ROMSize]uint8
	RAM [defines.RAMSize]uint8

	registers [defines.RegisterCount]uint8
	stack     *stack
	flags     *flags

	// program counter
	pc uint16

	// syntax and opcodes
	syn *asm.Syntax
}

func NewCPU(syn *asm.Syntax) *CPU {
	return &CPU{
		ROM:       [defines.ROMSize]uint8{},
		RAM:       [defines.RAMSize]uint8{},
		registers: [defines.RegisterCount]uint8{},
		stack:     newStack(defines.StackDepth),
		flags:     &flags{},
		pc:        0,
		syn:       syn,
	}
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
		fmt.Printf("instruction %s at PC = %d (0x%02x)\n", nextInstruction.name, cpu.pc, cpu.pc)

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

		// dump CPU state
		fmt.Println("======== CPU state ========")
		fmt.Printf("PC = %d\n", cpu.pc)
		fmt.Printf("flags:\n  %v\n", cpu.flags)
		fmt.Printf("registers:\n  ")
		for i := 0; i < defines.RegisterCount; i++ {
			fmt.Printf("r%d = %02x", i, cpu.registers[i])
			if i+1 != defines.RegisterCount {
				fmt.Printf(" | ")
			}
		}
		fmt.Println()
		fmt.Printf("===========================\n\n")

		if cpu.flags.halt {
			return
		}
	}
}
