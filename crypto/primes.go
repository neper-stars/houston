package crypto

/*
 * The first 64 prime numbers, after '2' (so all are odd). These are used
 * as starting seeds to the random number generator.
 *
 * IMPORTANT:  One number here is not prime (279).  I thought it should be
 * replaced with 269, which is prime.  StarsHostEditor 0.3 decompiled source
 * uses 279, and it turns out that an analysis of the stars EXE with a hex editor
 * also shows a primes table with 279.  Fun!
 */
var primes = []int{
	3, 5, 7, 11, 13, 17, 19, 23,
	29, 31, 37, 41, 43, 47, 53, 59,
	61, 67, 71, 73, 79, 83, 89, 97,
	101, 103, 107, 109, 113, 127, 131, 137,
	139, 149, 151, 157, 163, 167, 173, 179,
	181, 191, 193, 197, 199, 211, 223, 227,
	229, 233, 239, 241, 251, 257, 263, 279,
	271, 277, 281, 283, 293, 307, 311, 313,
}
