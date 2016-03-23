#define SYS_EXIT_GROUP	94

// _rt0_riscv_linux is the entry point.
TEXT _rt0_riscv_linux(SB),0,$0

	MOV	$SYS_EXIT_GROUP, A7
	MOV	$42, A0	// exit code
	SCALL
	RET
