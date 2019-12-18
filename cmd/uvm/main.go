package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sshaman1101/uvm/asm"
	"github.com/sshaman1101/uvm/cpu"
	"github.com/sshaman1101/uvm/defines"
)

var syntaxFile = flag.String("syntax", "syntax.yaml", "path to syntax definition")

var usageFunc = func() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage %s <rom.bin>\n", os.Args[0])
	flag.PrintDefaults()
}

func init() {
	flag.Usage = usageFunc
	flag.Parse()
}

func main() {
	if len(os.Args) < 2 {
		usageFunc()
		os.Exit(1)
	}

	syntax := asm.LoadSyntax(*syntaxFile)

	romFile := os.Args[1]
	image, err := ioutil.ReadFile(romFile)
	if err != nil {
		fmt.Printf("ERR: Failed to load ROM file from %s: %v", romFile, err)
		os.Exit(1)
	}

	if len(image) > defines.ROMSize {
		fmt.Printf("WARN: ROM image does not fits into memory "+
			"(size = %d, but %d bytes available).\n"+
			"Image will be truncated.\n", len(image), defines.ROMSize)
	}

	var rom = [defines.ROMSize]uint8{}
	copy(rom[:], image)

	uCPU := cpu.NewCPU(&syntax)
	uCPU.ROM = rom
	uCPU.Run()
}
