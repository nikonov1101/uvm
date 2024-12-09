package cpu

import (
	"fmt"
	"strings"

	colors "github.com/nikonov1101/colors.go"
	"github.com/sshaman1101/uvm/asm"
	"github.com/sshaman1101/uvm/defines"
)

type flags struct {
	zero  bool
	carry bool
	halt  bool
}

func (f *flags) String() string {
	return fmt.Sprintf("Z: %v | C: %v | H: %v", f.zero, f.carry, f.halt)
}

type CPU struct {
	flags *flags
	// general-purpose registers are not memory-mapped (yet).
	generalPurposeReg [defines.RegisterCount]uint8
	// pc actually 24 bits wide
	pc uint32
	// stack pointer, 24 bits wide as well
	sp uint32

	// address is 24 bit wide, first 8 bits are pointed by segmentSelectorReg
	// (inspired by 8088), next 16 bits are pointed by the address operand of
	// an instruction, thus:
	// MOV r1 $abcd
	// moves value at address sreg+0xabcd into r1.
	segmentSelectorReg uint8

	// all addressable memory
	mem [1 << defines.AddressWidth]uint8
}

const (
	startSegment = 0
	startAddress = 0
)

func NewCPU() *CPU {
	return &CPU{
		flags:              &flags{},
		sp:                 defines.StackInitialAddr, // very end of the memory
		pc:                 startSegment,
		segmentSelectorReg: startSegment,
	}
}

func (cpu *CPU) LoadROM(rom []byte) {
	if len(rom) == 0 {
		panic("empty ROM given")
	}

	romStart := 0x00FFFFFF & (uint32(startSegment)<<16 | uint32(startAddress))
	copy(cpu.mem[romStart:], rom)
}

func (cpu *CPU) Run() {
	for {
		// load next value from mem,
		// must be an instruction
		v := cpu.mem[cpu.pc]

		// STAGE 1: decode instruction
		// note: panics on invalid input
		next := cpu.decodeInstruction(v)
		var pcOffset uint32 = 0

		// STAGE 2: fetch the operands from memory
		for i := range next.operands {
			// calculate next address
			pcOffset++

			// fetch next operand from the memory
			given := cpu.mem[cpu.pc+pcOffset]
			expected := next.operands[i]

			// sanity check
			opName := checkOperand(given, expected.opType)
			// XXX debug
			fmt.Printf("  operand %s loaded\n", opName)

			// store operand *data* within instruction
			next.operands[i].value = given
		}

		// XXX print instruction with operators loaded
		fmt.Printf("PC: %04X :: %v\n", cpu.pc, next)

		// update PC with a number operands fetched,
		// do this before the actual execution, so
		// JUMP instructions may override the PC
		cpu.pc += pcOffset + 1

		// STAGE 3: execute the instruction
		cpu.execute(next)

		// XXX debug state on the fly
		cpu.debug()

		if cpu.flags.halt {
			return
		}
	}
}

// decodeInstruction checks that given opcode exists,
// if so, annotates it with desired operand types
// and the instruction name (just for the debug purposes).
func (cpu *CPU) decodeInstruction(opcode uint8) instruction {
	var operandsForOpCode []asm.OperandType
	var ok bool
	var mnemonic string

	// we'd like to have a mnemonic for given opcode,
	// so walk through the whole syntax definition.
	for name, opcodes := range asm.Syntax {
		// does this mnemonic implements given opcode?
		operandsForOpCode, ok = opcodes[opcode]
		if !ok {
			continue
		}
		mnemonic = name
		break
	}

	if !ok {
		panic(fmt.Sprintf("invalid instruiction %2x", opcode))
	}

	var instructionOperands []operand
	for _, op := range operandsForOpCode {
		// note: just a dirty crutch to add two address bytes for instruction.
		// Need to find a smarter way to handle such situation.
		if op == asm.OperandAddr {
			// XXX why???
			instructionOperands = append(instructionOperands, operand{opType: asm.OperandAddr}, operand{opType: asm.OperandAddr})
		} else {
			instructionOperands = append(instructionOperands, operand{opType: op})
		}
	}

	return instruction{
		name:     mnemonic,
		opCode:   opcode,
		operands: instructionOperands,
	}
}

func (cpu *CPU) debug() {
	var regs []string
	for i, r := range cpu.generalPurposeReg {
		rs := fmt.Sprintf("0x%02X", r)
		if r > 0 {
			rs = colors.Yellow(rs)
		}
		regs = append(regs, fmt.Sprintf("r%d: %s", i, rs))
	}

	zs := "Z: false"
	cs := "C: false"
	hs := "H: false"
	if cpu.flags.zero {
		zs = fmt.Sprintf("Z: %s", colors.Green("TRUE"))
	}
	if cpu.flags.carry {
		cs = fmt.Sprintf("C: %s", colors.Green("TRUE"))
	}
	if cpu.flags.halt {
		hs = colors.Red("HALT: TRUE")
	}

	flags := fmt.Sprintf("%s | %s | %s", zs, cs, hs)
	pc := colors.Cyan(fmt.Sprintf("0x%04X", cpu.pc))

	fmt.Printf("\tnext pc: %s | flags: %s\n", pc, flags)
	fmt.Printf("\t%s\n", strings.Join(regs, " "))
	fmt.Println("=============================================")
}
