package main

import (
	"github.com/sshaman1101/uvm/cpu"
)

var prog = [cpu.ROMSize]uint8{
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
	uCPU := cpu.NewCPU()
	uCPU.ROM = prog
	uCPU.Run()
}
