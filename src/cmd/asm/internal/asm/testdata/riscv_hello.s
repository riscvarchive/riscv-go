#define SYS_EXIT_GROUP	94
#define SYS_WRITE	64

// _rt0_riscv_linux is the entry point.
TEXT _rt0_riscv_linux(SB),0,$8
	// Write "H" to stdout...not quite to hello world, yet
	MOV	$72, T0 // 'H'
	MOVW	T0, 0(SP)
	MOV	SP, A1 // ptr to data
	MOV	$1, A0 // fd 1 for stdout
	MOV	$1, A2 // len("H") == 1
	MOV	$SYS_WRITE, A7
	SCALL
	// A0 is return value from syscall, convert to 0/1 for use as exit code and
	// put back in A0 for next syscall.
	// Note that, as the spec observes, SLTIU rd, rs1, 1 == SEQZ rd, rs1
	// TODO: Add SEQZ support directly
	SLTIU	A0, $1, A0
	MOV	$SYS_EXIT_GROUP, A7
	ECALL
	RET
