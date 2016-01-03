# RISC-V Go

## Quick Start

```sh
$ export GOROOT_BOOTSRAP=/path/to/prebuilt/go/tree
$ export GOROOT=$(pwd)
$ cd src
$ ./make.bash
$ GOARCH=riscv ../bin/go tool asm /path/to/some/riscv/asm.s
```
