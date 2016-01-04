#include <stdint.h>
#include <stdio.h>

/*
 * Build:
 * riscv64-unknown-elf-gcc -o /tmp/memcpy riscv-gas-test.c riscv-gas.s
 *
 * Test:
 * spike pk /tmp/memcpy
 */

extern int memcpy_c(void *dst, void *src, uint64_t size);

int main(void) {
	char a[] = "hello world";
	char b[sizeof(a)] = { '\0' };

	memcpy_c(b, a, sizeof(a));

	for (int i = 0; i < sizeof(a); i++) {
		if (a[i] != b[i]) {
			printf("i = %d got %c want %c\n", i, b[i], a[i]);
			return 1;
		}
	}
}
