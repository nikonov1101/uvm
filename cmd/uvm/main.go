package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sshaman1101/uvm/cpu"
)

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

	romFile := os.Args[1]
	image, err := os.ReadFile(romFile)
	if err != nil {
		fmt.Printf("ERR: Failed to load ROM file from %s: %v", romFile, err)
		os.Exit(1)
	}

	romSize := 1 << 16
	if len(image) > romSize {
		fmt.Printf("WARN: ROM image does not fits into memory "+
			"(size = %d, but %d bytes available).\n"+
			"Image will be truncated.\n", len(image), romSize)
	}

	uCPU := cpu.NewCPU()
	uCPU.LoadROM(image)
	uCPU.Run()
}
