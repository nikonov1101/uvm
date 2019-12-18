package main

import (
	"fmt"
	"os"

	"github.com/sshaman1101/uvm/asm"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage %s src.asm output.bin\n", os.Args[0])
		os.Exit(1)
	}

	in, out := os.Args[1], os.Args[2]
	fmt.Printf("Using %s as input file, store results in %s\n", in, out)

	src, err := os.OpenFile(in, os.O_RDONLY, 0600)
	if err != nil {
		fmt.Printf("Failed to read input from %s: %v\n", in, err)
		os.Exit(1)
	}

	mem := asm.Compile(src)
	dst, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Printf("Failed to open output file %s: %v\n", out, err)
		os.Exit(1)
	}

	if _, err := dst.Write(mem[:]); err != nil {
		fmt.Printf("Failed to write results to %s: %v\n", out, err)
		_ = os.Remove(out)
	}
}
