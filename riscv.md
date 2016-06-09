# RISC-V Go

## Quick Start

Setup:

```sh
$ export GOROOT_BOOTSRAP=/path/to/prebuilt/go/tree
$ export GOROOT=$(pwd)
$ cd src
$ ./make.bash
$ export PATH="$PATH:$GOROOT/misc/riscv"
```

Compile and run in spike using pk:

```sh
$ GOARCH=riscv GOOS=linux go run $GOROOT/riscv/riscvtest/hellomain.go; echo $?
```

Assemble and link:

```sh
$ GOARCH=riscv GOOS=linux $GOROOT/bin/go tool asm /path/to/some/riscv/asm.s
$ GOARCH=riscv GOOS=linux $GOROOT/bin/go tool link asm.o
```
