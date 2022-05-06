# Asm: simple assembler/disassembler

This tool is useful for assembling and disassembling individual instructions at
the command-line. It invokes a specified GNU toolchain to perform the
assembling/disassembling.

# Usage


```
Usage:
  asm [OPTIONS] INSTRUCTION

Application Options:
      --aa=      Additional arguments for the assembler, may be passed multiple times
      --la=      Additional arguments for the linker, may be passed multiple times
  -m, --machine= Set machine type for disassembly
  -d, --disas    Disassemble
  -p, --prefix=  Set GNU toolchain prefix (default: arm-none-eabi)
  -V, --verbose  Show additional information while running
  -v, --version  Show version number
  -h, --help     Show this help message
```

Examples:

```
$ asm "ldr pc, [pc, #4]"
e59ff004
```

```
$ asm -d e59ff004

/tmp/asm1313556893/insts.bin:     file format binary


Disassembly of section .data:

00000000 <.data>:
   0:	e59ff004 	ldr	pc, [pc, #4]	; 0xc
```

```
$ asm -p riscv64-unknown-elf --aa=-march=rv32i --la=-melf32lriscv "lw t0, 4(sp)"
412283
```
