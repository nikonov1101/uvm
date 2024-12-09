package asm

const (
	OpNOP   = 0x00
	OpPUSH  = 0x02
	OpPOP   = 0x03
	OpCLEAR = 0x04
	OpINC   = 0x05
	OpHALT  = 0x09

	OpCP  = 0x0A
	OpCPI = 0x0B

	OpJUMP      = 0x50
	OpJUMPIF_EQ = 0x51
	OpJUMPIF_NE = 0x52

	OpADDRegReg = 0x10
	OpADDRegVal = 0x11

	OpMOVRegReg = 0x20
	OpMOVRegVal = 0x21

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
	"PUSH":  {OpPUSH: {OperandReg}},
	"POP":   {OpPOP: {OperandReg}},
	"CLEAR": {OpCLEAR: {OperandReg}},
	"INC":   {OpINC: {OperandReg}},
	"HALT":  {OpHALT: {}},

	"CP": {
		OpCP:  {OperandReg, OperandReg},
		OpCPI: {OperandReg, OperandValue},
	},

	"JUMP": {OpJUMP: {OperandAddr}},
	"JE":   {OpJUMPIF_EQ: {OperandAddr}},
	"JNE":  {OpJUMPIF_NE: {OperandAddr}},

	"ADD": {
		OpADDRegReg: {OperandReg, OperandReg},   //  do reg1 + reg2 and store the result in reg1
		OpADDRegVal: {OperandReg, OperandValue}, // do reg1 + value and store the result in reg1
	},
	"MOV": {
		OpMOVRegReg: {OperandReg, OperandReg},   // move reg2' value to reg1
		OpMOVRegVal: {OperandReg, OperandValue}, //  move value to reg1 immediately
	},

	// external memory
	"LOAD":  {OpLOAD: {OperandReg, OperandAddr}},  // load register with a value stored at addr
	"STORE": {OpSTORE: {OperandAddr, OperandReg}}, // store reg's value at address
}
