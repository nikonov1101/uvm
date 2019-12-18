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

