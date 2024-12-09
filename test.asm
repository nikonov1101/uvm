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
LOAD r3 $1a2b

MOV r2 #aa
STORE $1a2c r2

;test conditionals
MOV r3 #1
MOV r4 #a1
CP r3 r4
JNE $001A

NOP
NOP
NOP
NOP
NOP

.text $001A
INC r3
INC r3
CP r3 #3
JE $01FF


NOP
; check jump
JUMP $00FF

; place more instructions at $00ff
; check that we can compile .text's
.text $00FF
HALT

.text $01FF
    HALT

.text $02FF
    HALT

; place random value at $0100
; check that we can compile .byte's
.byte $0101 #42

