package main

import (
	"github.com/sshaman1101/uvm/asm"
	"github.com/sshaman1101/uvm/cpu"
)

var byteCode = [cpu.ROMSize]uint8{
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

var asmCode = `
; check addition
MOV r0, #3
MOV r1, #2
ADD r0, r1
; can we do nothing?
NOP
NOP
; check push
PUSH r1
NOP
NOP
; check pop
POP r5
; check mem load
MOV r3, $0101
NOP
; check jump
JUMP $00FF

; place more instructions at $00ff
; check that we can compile .text's
.text $00FF
HALT

; place random value at $0100
; check that we can compile .byte's
.byte $0101 #42
`

func main() {
	p := asm.Compile(asmCode)

	uCPU := cpu.NewCPU()
	uCPU.ROM = p
	uCPU.Run()
}
