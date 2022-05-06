package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jessevdk/go-flags"
)

// Version is the current version and should be inserted at compile-time by a
// linker flag.
var Version = "0.0.0-unknown"

var opts struct {
	AsArgs  []string `long:"aa" description:"Additional arguments for the assembler, may be passed multiple times"`
	LdArgs  []string `long:"la" description:"Additional arguments for the linker, may be passed multiple times"`
	Machine string   `short:"m" long:"machine" description:"Set machine type for disassembly"`
	Disas   bool     `short:"d" long:"disas" description:"Disassemble"`
	Prefix  string   `short:"p" long:"prefix" description:"Set GNU toolchain prefix" default:"arm-none-eabi"`
	Verbose bool     `short:"V" long:"verbose" description:"Show additional information while running"`
	Version bool     `short:"v" long:"version" description:"Show version number"`
	Help    bool     `short:"h" long:"help" description:"Show this help message"`
}

var memmap = []byte(`
SECTIONS
{
    .text 0 :  { *(.text*) }
    .data : { *(.data*) } 
    .rodata : { *(.rodata*) }
    .bss : {
        *(.bss*)
		*(COMMON)
        . = ALIGN(8);
    }
}
`)

func docmd(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	if opts.Verbose {
		fmt.Println(cmd)
	}
	errb := &bytes.Buffer{}
	b := &bytes.Buffer{}
	cmd.Stderr = errb
	cmd.Stdout = b

	err := cmd.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, errb.String())
		os.Exit(1)
	}
	return b.String()
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func ffs(n uint32) int {
	for i := 0; i < 32; i++ {
		if (n & 1) != 0 {
			return i
		}
		n >>= 1
	}
	return 32
}

func assemble(insts string) []byte {
	dir, err := ioutil.TempDir("", "asm")
	must(err)

	asm := filepath.Join(dir, "insts.s")
	obj := filepath.Join(dir, "insts.o")
	elf := filepath.Join(dir, "insts.elf")
	bin := filepath.Join(dir, "insts.bin")
	ldscript := filepath.Join(dir, "memmap.ld")

	must(ioutil.WriteFile(asm, []byte(insts), 0666))
	must(ioutil.WriteFile(ldscript, memmap, 0666))

	as := fmt.Sprintf("%s-as", opts.Prefix)
	ld := fmt.Sprintf("%s-ld", opts.Prefix)
	objcopy := fmt.Sprintf("%s-objcopy", opts.Prefix)

	asargs := append([]string{"--no-warn", asm, "-o", obj}, opts.AsArgs...)
	docmd(as, asargs...)
	ldargs := append([]string{obj, "-T", ldscript, "-o", elf}, opts.LdArgs...)
	docmd(ld, ldargs...)
	docmd(objcopy, elf, "-O", "binary", bin)

	data, err := ioutil.ReadFile(bin)
	must(err)
	return data
}

func arch(triple string) string {
	if opts.Machine != "" {
		return opts.Machine
	}
	before, _, _ := strings.Cut(triple, "-")
	return before
}

func disassemble(mcode string) {
	insts := strings.Split(mcode, "\n")
	if len(insts) == 0 {
		return
	}
	dir, err := ioutil.TempDir("", "asm")
	must(err)

	asm := filepath.Join(dir, "insts.bin")
	buf := make([]byte, 4*len(insts))
	for _, inst := range insts {
		u, err := strconv.ParseUint(inst, 16, 32)
		must(err)
		binary.LittleEndian.PutUint32(buf, uint32(u))
	}
	must(ioutil.WriteFile(asm, buf, 0666))

	objdump := fmt.Sprintf("%s-objdump", opts.Prefix)

	out := docmd(objdump, "-b", "binary", "-m", arch(opts.Prefix), "-D", asm)
	fmt.Println(out)
}

func emitMcode(pattern string) {
	data := assemble(pattern)
	for len(data) >= 4 {
		fmt.Printf("%x\n", binary.LittleEndian.Uint32(data))
		data = data[4:]
	}
	if len(data) >= 2 {
		fmt.Printf("%x\n", binary.LittleEndian.Uint16(data))
		data = data[2:]
	}
	if len(data) > 0 {
		fmt.Printf("%d\n", data)
	}
}

func main() {
	flagparser := flags.NewParser(&opts, flags.PassDoubleDash|flags.PrintErrors)
	flagparser.Usage = "[OPTIONS] INSTRUCTION"
	args, err := flagparser.Parse()
	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		fmt.Println("asm version", Version)
		os.Exit(0)
	}

	if opts.Help {
		flagparser.WriteHelp(os.Stdout)
		os.Exit(0)
	}

	asm := ""
	if len(args) <= 0 {
		b, err := io.ReadAll(os.Stdin)
		must(err)
		asm = string(b)
	} else {
		asm = args[0]
	}

	if opts.Disas {
		disassemble(asm)
	} else {
		emitMcode(asm)
	}
}
