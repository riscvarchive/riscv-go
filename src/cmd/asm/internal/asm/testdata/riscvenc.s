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

	// This jump can get printed as JMP 2 because it goes to the second
	// instruction in the function.  (The first instruction is an invisible
	// stack pointer adjustment.)
	JMP	start		// JMP	2	// 6ff01ffb

	RDCYCLE	T0				// f32200c0
