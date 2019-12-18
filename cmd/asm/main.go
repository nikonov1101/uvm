package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sshaman1101/uvm/asm"
)

var syntaxFile = flag.String("syntax", "syntax.yaml", "path to syntax definition")

var usageFunc = func() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage %s <src.asm> <output.bin>\n", os.Args[0])
	flag.PrintDefaults()
}

func init() {
	flag.Usage = usageFunc
	flag.Parse()
}

func main() {
	if len(os.Args) < 3 {
		usageFunc()
		os.Exit(1)
	}

	syn := asm.LoadSyntax(*syntaxFile)

	in, out := os.Args[1], os.Args[2]
	fmt.Printf("Using %s as input file, store results in %s\n", in, out)

	src, err := os.OpenFile(in, os.O_RDONLY, 0600)
	if err != nil {
		fmt.Printf("Failed to read input from %s: %v\n", in, err)
		os.Exit(1)
	}

	mem := asm.Compile(src, &syn)
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
