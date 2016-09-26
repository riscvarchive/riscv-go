package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

var tests = [...]struct {
	name string
	want int
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
	{name: "fmv", want: 5},
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
}

func main() {
	var failed bool
	spike, err := exec.LookPath("spike")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove("tmp")
	for _, test := range tests {
		filename := test.name + ".go"
		// build
		cmd := exec.Command("go", "build", "-o", "tmp", filename)
		cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=riscv")
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("compilation of %q failed:\n%s\n", filename, out)
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
			fmt.Printf("rc(%q)=0, want %d\n", filename, test.want)
			failed = true
			continue
		}
		ee, ok := err.(*exec.ExitError)
		if !ok {
			log.Printf("%q: unexpected execution error type %T: %v", filename, err, err)
			failed = true
			continue
		}
		ws := ee.ProcessState.Sys().(syscall.WaitStatus)
		rc := ws.ExitStatus()
		if rc != test.want {
			fmt.Printf("rc(%q)=%d, want %d\n", filename, rc, test.want)
			failed = true
			continue
		}
	}

	if failed {
		os.Exit(1)
	}
}
