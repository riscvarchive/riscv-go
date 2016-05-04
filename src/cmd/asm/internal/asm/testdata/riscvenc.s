TEXT asmtest(SB),7,$0
start:
	ADD	T1, T0, T2			// b3836200
	ADD	T0, T1				// 33035300
	ADD	$2047, T0, T1			// 1383f27f
	ADD	$-2048, T0, T1			// 13830280
	ADD	$2047, T0			// 9382f27f
	ADD	$-2048, T0			// 93820280

	SUB	T1, T0, T2			// b3836240
	SUB	T0, T1				// 33035340

	SLL	T1, T0, T2			// b3936200
	SLL	T0, T1				// 33135300
	SLL	$1, T0, T1			// 13931200
	SLL	$1, T0				// 93921200
	SRL	T1, T0, T2			// b3d36200
	SRL	T0, T1				// 33535300
	SRL	$1, T0, T1			// 13d31200
	SRL	$1, T0				// 93d21200
	SRA	T1, T0, T2			// b3d36240
	SRA	T0, T1				// 33535340
	SRA	$1, T0, T1			// 13d31200
	SRA	$1, T0				// 93d21200

	AND	T1, T0, T2			// b3f36200
	AND	T0, T1				// 33735300
	AND	$1, T0, T1			// 13f31200
	AND	$1, T0				// 93f21200
	OR	T1, T0, T2			// b3e36200
	OR	T0, T1				// 33635300
	OR	$1, T0, T1			// 13e31200
	OR	$1, T0				// 93e21200
	XOR	T1, T0, T2			// b3c36200
	XOR	T0, T1				// 33435300
	XOR	$1, T0, T1			// 13c31200
	XOR	$1, T0				// 93c21200

	// These jumps can get printed as jumps to 2 because they go to the
	// second instruction in the function.  (The first instruction is an
	// invisible stack pointer adjustment.)
	JMP	start		// JMP	2	// 6ff01ff8
	JAL	T0, start	// JAL T0, 2	// eff2dff7
	BEQ	T0, T1, start	// BEQ T0, T1, 2	// e38c62f6
	BNE	T0, T1, start	// BNE T0, T1, 2	// e39a62f6
	BLT	T0, T1, start	// BLT T0, T1, 2	// e3c862f6
	BGE	T0, T1, start	// BGE T0, T1, 2	// e3d662f6
	BLTU	T0, T1, start	// BLTU T0, T1, 2	// e3e462f6
	BGEU	T0, T1, start	// BGEU T0, T1, 2	// e3f262f6

	// Ditto.  (This large constant load expands into a lui+addi.)
	MOV	$54321, T0			// b7d20000

	JMP	(T0)				// 67800200
	JMP	4(T0)				// 67804200

	// Jump to T0, link address in T1.
	JALR	T1, (T0)			// 67830200
	JALR	T1, 4(T0)			// 67834200

	RET					// 67800000

	SCALL					// 73000000
	RDCYCLE	T0				// f32200c0
	RDTIME	T0				// f32210c0
	RDINSTRET	T0			// f32220c0

	AUIPC	$0, A0 				// 17050000
	AUIPC	$0, A1 				// 97050000
	AUIPC	$1, A0				// 17150000

	LUI	$167, A5			// b7770a00

	MOV	T0, T1				// 13830200
	MOV	$2047, T0			// 9302f07f
	MOV	$-2048, T0			// 93020080

	MOVB	(T0), T1			// 03830200
	MOVB	4(T0), T1			// 03834200
	MOVH	(T0), T1			// 03930200
	MOVH	4(T0), T1			// 03934200
	MOVW	(T0), T1			// 03a30200
	MOVW	4(T0), T1			// 03a34200
	MOV	(T0), T1			// 03b30200
	MOV	4(T0), T1			// 03b34200
	MOVB	T0, (T1)			// 23005300
	MOVB	T0, 4(T1)			// 23025300
	MOVH	T0, (T1)			// 23105300
	MOVH	T0, 4(T1)			// 23125300
	MOVW	T0, (T1)			// 23205300
	MOVW	T0, 4(T1)			// 23225300
	MOV	T0, (T1)			// 23305300
	MOV	T0, 4(T1)			// 23325300

	SLT	T0, T1, T2			// b3a36200
	SLT	T0, $55, T2			// 93a37203
	SLTU	T0, T1, T2			// b3b36200
	SLTU	T0, $55, T2			// 93b37203

	SEQZ	A5, A5				// 93b71700
	SNEZ	A5, A5				// b337f000
