.text

.global memcpy
memcpy:
	/* dst */
	ld	t0, (sp)
	/* src */
	ld	t1, 8(sp)
	/* size */
	ld	t2, 16(sp)

loop:
	bge	zero, t2, done

	lbu	t4, (t1)
	sb	t4, (t0)

	addi	t0, t0, 1
	addi	t1, t1, 1
	addi	t2, t2, -1

	j	loop

done:
	ret

/*
 * Convert C calling convention to Go calling convention.
 *
 * a0 = dst, a1 = src, a2 = size
 */
.global memcpy_c
memcpy_c:
	addi	sp, sp, -32
	sd	a0, (sp)
	sd	a1, 8(sp)
	sd	a2, 16(sp)

	sd	ra, 24(sp)

	jal	memcpy

	ld	ra, 24(sp)

	addi	sp, sp, 32

	ret
