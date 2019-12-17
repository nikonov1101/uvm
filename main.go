package main

import (
	"github.com/sshaman1101/uvm/cpu"
)

var prog = [cpu.ROMSize]uint8{
	0x21, 0x00, 0x03, // mov r0, #3,
	0x21, 0x01, 0x02, // mov r1, #2
	0x10, 0x00, 0x01, // add r0, r1 (store result in r1)
	0x00, 0x00, // NOP, NOP
	0x02, 0x00, // PUSH r1
	0x00, 0x00, // NOP, NOP
	0x03, 0x05, // POP r5
	0x22, 0x3, 0x19, 0x00, // mov r3, $0015
	0x00,             // NOP
	0x01, 0xFF, 0x00, // JUMP to $00FF (see prog modification in main() func)
	0xaa, // just a value at addr = 0x19
}

// all opCodes are NOPs, with HALT at end (see main() func)
var nopProg = [cpu.ROMSize]uint8{}

func main() {
	// monkey-patching be like
	prog[255] = 0x09              // HALT
	nopProg[cpu.ROMSize-1] = 0x09 // HALT

	uCPU := cpu.NewCPU()
	uCPU.ROM = nopProg
	uCPU.Run()
}
