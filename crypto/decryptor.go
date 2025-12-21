package crypto

// StarsRandom is a pseudo-random number generator used by Stars!
type StarsRandom struct {
	seedA  int
	seedB  int
	rounds int
}

// NewStarsRandom creates a new StarsRandom with the given seeds and initial rounds
func NewStarsRandom(seed1, seed2, initRounds int) *StarsRandom {
	random := &StarsRandom{
		seedA:  seed1,
		seedB:  seed2,
		rounds: initRounds,
	}

	for i := 0; i < initRounds; i++ {
		random.NextRandom()
	}

	return random
}

// NextRandom generates the next random number in the sequence
func (r *StarsRandom) NextRandom() int {
	seedApartA := (r.seedA % 53668) * 40014
	seedApartB := (r.seedA / 53668) * 12211
	newSeedA := seedApartA - seedApartB

	seedBpartA := (r.seedB % 52774) * 40692
	seedBpartB := (r.seedB / 52774) * 3791
	newSeedB := seedBpartA - seedBpartB

	if newSeedA < 0 {
		newSeedA += 0x7fffffab
	}

	if newSeedB < 0 {
		newSeedB += 0x7fffff07
	}

	r.seedA = newSeedA
	r.seedB = newSeedB

	randomNumber := r.seedA - r.seedB
	if r.seedA < r.seedB {
		randomNumber += 0x100000000
	}

	return randomNumber
}

// Decryptor handles decryption of Stars! file data
type Decryptor struct {
	random *StarsRandom
}

// NewDecryptor creates a new Decryptor instance
func NewDecryptor() *Decryptor {
	return &Decryptor{}
}

// InitDecryption initializes the decryptor with game parameters
func (d *Decryptor) InitDecryption(salt, gameId, turn, playerIndex, shareware int) {
	// Use two prime numbers as random seeds.
	// First one comes from the lower 5 bits of the salt
	index1 := salt & 0x1F
	// Second index comes from the next higher 5 bits
	index2 := (salt >> 5) & 0x1F

	// Adjust our indexes if the highest bit (bit 11) is set
	// If set, change index1 to use the upper half of our primes table
	if (salt >> 10) == 1 {
		index1 += 32
	} else {
		// Else index2 uses the upper half of the primes table
		index2 += 32
	}

	// Determine the number of initialization rounds from 4 other data points
	// 0 or 1 if shareware (I think this is correct, but may not be - so far
	// I have not encountered a shareware flag)
	part1 := shareware
	// Lower 2 bits of player number, plus 1
	part2 := (playerIndex & 0x3) + 1
	// Lower 2 bits of turn number, plus 1
	part3 := (turn & 0x3) + 1
	// Lower 2 bits of gameId, plus 1
	part4 := (gameId & 0x3) + 1
	// Now put them all together, this could conceivably generate up to 65
	// rounds  (4 * 4 * 4) + 1
	rounds := (part4 * part3 * part2) + part1

	// Now initialize our random number generator
	seed1 := primes[index1]
	seed2 := primes[index2]

	d.random = NewStarsRandom(seed1, seed2, rounds)
}

// DecryptBytes decrypts a byte slice using the initialized random generator
func (d *Decryptor) DecryptBytes(b []byte) []byte {
	byteArray := make([]byte, len(b))
	copy(byteArray, b)
	// Add padding to 4 bytes
	size := len(byteArray)
	paddedSize := (size + 3) &^ 3 // Round up to the nearest multiple of 4
	padding := paddedSize - size

	for i := 0; i < padding; i++ {
		byteArray = append(byteArray, 0x00)
	}

	decryptedBytes := make([]byte, 0)

	// Now decrypt, processing 4 bytes at a time
	for i := 0; i < paddedSize; i += 4 {
		// Swap bytes using indexes in this order:  4 3 2 1
		chunk := (int(byteArray[i+3]) << 24) | (int(byteArray[i+2]) << 16) | (int(byteArray[i+1]) << 8) | int(byteArray[i])
		// XOR with a "random" number
		decryptedChunk := chunk ^ d.random.NextRandom()

		// Write out the decrypted data, swapped back
		decryptedBytes = append(decryptedBytes, byte(decryptedChunk&0xFF))
		decryptedBytes = append(decryptedBytes, byte((decryptedChunk>>8)&0xFF))
		decryptedBytes = append(decryptedBytes, byte((decryptedChunk>>16)&0xFF))
		decryptedBytes = append(decryptedBytes, byte((decryptedChunk>>24)&0xFF))
	}

	// Remove padding
	for i := 0; i < padding; i++ {
		byteArray = byteArray[:len(byteArray)-1]
		decryptedBytes = decryptedBytes[:len(decryptedBytes)-1]
	}

	return decryptedBytes
}
