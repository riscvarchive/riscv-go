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
	SRA	$1, T0, T1			// 13d31240
	SRA	$1, T0				// 93d21240

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

	JMP	(T0)				// 67800200
	JMP	4(T0)				// 67804200

	// Jump to T0, link address in T1.
	JALR	T1, (T0)			// 67830200
	JALR	T1, 4(T0)			// 67834200

	// Encoded as
	//	AUIPC $0, TMP
	//	ADDI $0, TMP
	//	JALR TMP
	// with a R_RISCV_PCREL_ITYPE relocation. The linker resolves the real
	// address.
	CALL	asmtest(SB)			// 970f0000
	// Encoded as
	//	AUIPC $0, TMP
	//	ADDI $0, TMP
	//	JMP TMP
	// with a R_RISCV_PCREL_ITYPE relocation. The linker resolves the real
	// address.
	JMP	asmtest(SB)			// 970f0000

	ECALL					// 73000000
	SCALL					// 73000000
	RDCYCLE	T0				// f32200c0
	RDTIME	T0				// f32210c0
	RDINSTRET	T0			// f32220c0

	AUIPC	$0, A0 				// 17050000
	AUIPC	$0, A1 				// 97050000
	AUIPC	$1, A0				// 17150000

	LUI	$167, A5			// b7770a00

	MOV	T0, T1				// 13830200
	MOV	$2047, T0			// 9b02f07f
	MOV	$-2048, T0			// 9b020080

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

	SLT	T1, T0, T2			// b3a36200
	SLT	$55, T0, T2			// 93a37203
	SLTU	T1, T0, T2			// b3b36200
	SLTU	$55, T0, T2			// 93b37203

	SEQZ	A5, A5				// 93b71700
	SNEZ	A5, A5				// b337f000

	// Arbitrary bytes (entered in little-endian mode)
	WORD	$0x12345678	// WORD $305419896	// 78563412
	WORD	$0x9abcdef0	// WORD $2596069104	// f0debc9a


	// M extension
	MUL	T0, T1, T2			// b3035302
	MULH	T0, T1, T2			// b3135302
	MULHU	T0, T1, T2			// b3335302
	MULHSU	T0, T1, T2			// b3235302
	MULW	T0, T1, T2			// bb035302

	DIV	T0, T1, T2			// b3435302
	DIVU	T0, T1, T2			// b3535302
	REM	T0, T1, T2			// b3635302
	REMU	T0, T1, T2			// b3735302
	DIVW	T0, T1, T2			// bb435302
	DIVUW	T0, T1, T2			// bb535302
	REMW	T0, T1, T2			// bb635302
	REMUW	T0, T1, T2			// bb735302


	// F extension
	FADDS	FT1, FT0, FT2			// 53011000
	FSUBS	FT1, FT0, FT2			// 53011008
	FMULS	FT1, FT0, FT2			// 53011010
	FDIVS	FT1, FT0, FT2			// 53011018
	FSQRTS	FT0, FT1			// d3000058
	FNEGS	FT0, FT1			// d3100020
	FSGNJS	FT1, FT0, FT2			// 53011020
	FSGNJNS	FT1, FT0, FT2			// 53111020
	FSGNJXS	FT1, FT0, FT2			// 53211020
	FCVTSW	T0, FT0				// 538002d0
	FCVTSL	T0, FT0				// 538022d0
	FCVTWS	FT0, T0				// d31200c0
	FCVTLS	FT0, T0				// d31220c0
	MOVF	4(T0), FT0			// 07a04200
	MOVF	FT0, 4(T0)			// 27a20200
	MOVF	FT0, FT1			// d3000020
	FEQS	FT0, FT1, T2			// d3a300a0
	FNES	FT0, FT1, T2			// d3a300a0
	FLTS	FT0, FT1, T2			// d39300a0
	FLES	FT0, FT1, T2			// d38300a0


	// D extension
	FADDD	FT1, FT0, FT2			// 53011002
	FSUBD	FT1, FT0, FT2			// 5301100a
	FMULD	FT1, FT0, FT2			// 53011012
	FDIVD	FT1, FT0, FT2			// 5301101a
	FSQRTD	FT0, FT1			// d300005a
	FNEGD	FT0, FT1			// d3100022
	FSGNJD	FT1, FT0, FT2			// 53011022
	FSGNJND	FT1, FT0, FT2			// 53111022
	FSGNJXD	FT1, FT0, FT2			// 53211022
	FCVTDW	T0, FT0				// 538002d2
	FCVTDL	T0, FT0				// 538022d2
	FCVTWD	FT0, T0				// d31200c2
	FCVTLD	FT0, T0				// d31220c2
	MOVD	4(T0), FT0			// 07b04200
	MOVD	FT0, 4(T0)			// 27b20200
	MOVD	FT0, FT1			// d3000022
	FEQD	FT0, FT1, T0			// d3a200a2
	FNED	FT0, FT1, T0			// d3a200a2
	FLTD	FT0, FT1, T0			// d39200a2
	FLED	FT0, FT1, T0			// d38200a2
