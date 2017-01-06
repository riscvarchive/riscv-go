# The Go Programming Language

## RISC-V Go Port

This repository is home of the RISC-V port of the Go programming language.

The upstream Go project can be found at https://github.com/golang/go.

### Quick Start

Setup:

```sh
$ export GOROOT_BOOTSTRAP=/path/to/prebuilt/go/tree
$ export GOROOT=$(pwd)
$ cd src
$ ./make.bash
$ export PATH="$PATH:$GOROOT/misc/riscv"
$ export PATH="$PATH:$GOROOT/bin"
```

Compile and run in spike using pk:

```sh
$ GOARCH=riscv GOOS=linux go run $GOROOT/riscvtest/add.go
```

Build:

```sh
$ GOARCH=riscv GOOS=linux go build $GOROOT/riscvtest/add.go
```

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
