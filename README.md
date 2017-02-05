# The Go Programming Language

## RISC-V Go Port

This repository is home of the RISC-V port of the Go programming language.

The upstream Go project can be found at https://github.com/golang/go.

### Quick Start

Setup:

```sh
$ git clone https://review.gerrithub.io/riscv/riscv-go riscv-go
$ cd riscv-go
$ git checkout riscvdev  # RISC-V work happens on this branch
$ export GOROOT_BOOTSTRAP=/path/to/prebuilt/go/tree
$ export PATH="$(pwd)/misc/riscv:$(pwd)/bin:$PATH"
$ cd src
$ ./make.bash
```

Compile and run in spike using pk (which are expected to be in PATH):

```sh
$ GOARCH=riscv GOOS=linux go run ../riscvtest/add.go
```

Build:

```sh
$ GOARCH=riscv GOOS=linux go build ../riscvtest/add.go
```

Test:

Our basic tests are in the `riscvtest` directory:

```sh
$ cd ../riscvtest
$ go run run.go
```

If this exits without error, all is well!

Note that these tests currently use the special builtin `riscvexit` to exit,
until we can build the standard library and use os.Exit.

### QEMU

Spike plus pk support only a small subset of Linux syscalls and will not be
capable of supporting the full Go runtime.

The [RISC-V QEMU port](https://github.com/riscv/riscv-qemu) supports a much
wider set of syscalls with its "User Mode Simulation". See [Method
2](https://github.com/riscv/riscv-qemu#method-2a-fedora-24-userland-with-user-mode-simulation-recommended)
in the QEMU README for instructions.

### Contributing

All contributors must sign the upstream [Contributor License
Agreement](https://golang.org/doc/contribute.html#cla), as this port will be
merged into upstream Go upon completion.

Code review occurs via our
[GerritHub](https://review.gerrithub.io/#/admin/projects/riscv/riscv-go)
project, rather than via GitHub Pull Requests.

The upstream [contribution guidelines](https://golang.org/doc/contribute.html)
include a basic overview of using Gerrit. While the upstream Go Gerrit server
is different from ours, `codereview.cfg` will configure `git-codereview` to
send CLs to GerritHub.
