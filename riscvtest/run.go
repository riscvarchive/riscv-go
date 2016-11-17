package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

var tests = [...]struct {
	name string
	want int
	dir  bool // test is a multi-file package in its own directory
}{
	{name: "hellomain", want: 42},
	{name: "loadstore", want: 0},
	{name: "add", want: 12},
	{name: "sub", want: 12},
	{name: "mul", want: 12},
	// {name: "div", want: 12}, // TODO: uncomment when we fix the runtime.panicdivide linker failure
	// {name: "rem", want: 12}, // TODO: uncomment when we fix the runtime.panicdivide linker failure
	{name: "fadd", want: 12},
	{name: "dadd", want: 12},
	{name: "fmv"},
	{name: "dmv", want: 5},
	{name: "zero8", want: 3},
	{name: "cmp"},
	{name: "fcmp"},
	{name: "dcmp"},
	{name: "bits"},
	{name: "ext"},
	{name: "bool"},
	{name: "nilcheck", want: 255}, // intentionally faults
	{name: "com"},
	{name: "left_shift"},
	{name: "right_shift"},
	{name: "right_shift_unsigned"},
	{name: "avg"},
	{name: "lrot"},
	{name: "call"},
	{name: "keepalive"},
	{name: "global"},
	{name: "live"},
	{name: "typeswitch"},
	{name: "immediate"},
	{name: "jmp", dir: true},
}

func main() {
	var failed bool
	spike, err := exec.LookPath("spike")
	if err != nil {
		log.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("unable to get working directory: %v", err)
	}

	tmp := filepath.Join(cwd, "tmp")
	defer os.Remove(tmp)
	for _, test := range tests {
		// build
		cmd := exec.Command("go", "build", "-o", tmp)
		cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=riscv")
		if test.dir {
			// Build everything in directory.
			cmd.Dir = filepath.Join(cwd, test.name)
		} else {
			// Build the file.
			filename := test.name + ".go"
			cmd.Args = append(cmd.Args, filename)
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("compilation of %q failed:\n%s\n", test.name, out)
			failed = true
			continue
		}

		// run
		cmd = exec.Command(spike, "pk", "tmp")
		out, err = cmd.CombinedOutput()
		if err == nil {
			if test.want == 0 {
				continue
			}
			fmt.Printf("rc(%q)=0, want %d\n", test.name, test.want)
			failed = true
			continue
		}
		ee, ok := err.(*exec.ExitError)
		if !ok {
			log.Printf("%q: unexpected execution error type %T: %v", test.name, err, err)
			failed = true
			continue
		}
		ws := ee.ProcessState.Sys().(syscall.WaitStatus)
		rc := ws.ExitStatus()
		if rc != test.want {
			fmt.Printf("rc(%q)=%d, want %d\n", test.name, rc, test.want)
			failed = true
			continue
		}
	}

	if failed {
		os.Exit(1)
	}
}
