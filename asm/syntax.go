package asm

const (
	OpNOP   = 0x00
	OpJUMP  = 0x01
	OpPUSH  = 0x02
	OpPOP   = 0x03
	OpCLEAR = 0x04
	OpINC   = 0x05
	OpHALT  = 0x09

	OpADDRegReg = 0x10
	OpADDRegVal = 0x11

	OpMOVRegReg = 0x20
	OpMOVRegVal = 0x21

	OpLPM   = 0x22
	OpLOAD  = 0x30
	OpSTORE = 0x40
)

/*
asm syntax help
	registers:    r0, r1, ... r7
	values (hex): #10, #0F, #123
	addr (hex):   $00, $FF
*/

// Syntax maps instruction name to opcodes with operands
var Syntax = map[string]map[uint8][]OperandType{
	"NOP":   {OpNOP: {}},
	"JUMP":  {OpJUMP: {OperandAddr}},
	"PUSH":  {OpPUSH: {OperandReg}},
	"POP":   {OpPOP: {OperandReg}},
	"CLEAR": {OpCLEAR: {OperandReg}},
	"INC":   {OpINC: {OperandReg}},
	"HALT":  {OpHALT: {}},

	"ADD": {
		OpADDRegReg: {OperandReg, OperandReg},   //  do reg1 + reg2 and store the result in reg1
		OpADDRegVal: {OperandReg, OperandValue}, // do reg1 + value and store the result in reg1
	},
	"MOV": {
		OpMOVRegReg: {OperandReg, OperandReg},   // move reg2' value to reg1
		OpMOVRegVal: {OperandReg, OperandValue}, //  move value to reg1 immediately
	},

	"LPM": {OpLPM: {OperandReg, OperandAddr}}, // load ROM value at given addr to reg

	// external memory
	"LOAD":  {OpLOAD: {OperandReg, OperandAddr}},  // load register with a value stored at addr
	"STORE": {OpSTORE: {OperandAddr, OperandReg}}, // store reg's value at address
}
