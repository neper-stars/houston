package crypto

// Encryptor handles encryption of Stars! file data.
// Since Stars! uses XOR encryption, encryption and decryption are the same operation.
type Encryptor struct {
	random *StarsRandom
}

// NewEncryptor creates a new Encryptor instance.
func NewEncryptor() *Encryptor {
	return &Encryptor{}
}

// InitEncryption initializes the encryptor with game parameters.
// This uses the same algorithm as decryption since XOR is symmetric.
func (e *Encryptor) InitEncryption(salt, gameId, turn, playerIndex, shareware int) {
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
	part1 := shareware
	part2 := (playerIndex & 0x3) + 1
	part3 := (turn & 0x3) + 1
	part4 := (gameId & 0x3) + 1
	rounds := (part4 * part3 * part2) + part1

	// Now initialize our random number generator
	seed1 := primes[index1]
	seed2 := primes[index2]

	e.random = NewStarsRandom(seed1, seed2, rounds)
}

// EncryptBytes encrypts a byte slice using the initialized random generator.
// Since XOR is symmetric, this uses the same algorithm as decryption.
func (e *Encryptor) EncryptBytes(b []byte) []byte {
	byteArray := make([]byte, len(b))
	copy(byteArray, b)
	// Add padding to 4 bytes
	size := len(byteArray)
	paddedSize := (size + 3) &^ 3 // Round up to the nearest multiple of 4
	padding := paddedSize - size

	for i := 0; i < padding; i++ {
		byteArray = append(byteArray, 0x00)
	}

	encryptedBytes := make([]byte, 0)

	// Now encrypt, processing 4 bytes at a time
	for i := 0; i < paddedSize; i += 4 {
		// Swap bytes using indexes in this order: 4 3 2 1
		chunk := (int(byteArray[i+3]) << 24) | (int(byteArray[i+2]) << 16) | (int(byteArray[i+1]) << 8) | int(byteArray[i])
		// XOR with a "random" number
		encryptedChunk := chunk ^ e.random.NextRandom()

		// Write out the encrypted data, swapped back
		encryptedBytes = append(encryptedBytes, byte(encryptedChunk&0xFF))
		encryptedBytes = append(encryptedBytes, byte((encryptedChunk>>8)&0xFF))
		encryptedBytes = append(encryptedBytes, byte((encryptedChunk>>16)&0xFF))
		encryptedBytes = append(encryptedBytes, byte((encryptedChunk>>24)&0xFF))
	}

	// Remove padding
	for i := 0; i < padding; i++ {
		byteArray = byteArray[:len(byteArray)-1]
		encryptedBytes = encryptedBytes[:len(encryptedBytes)-1]
	}

	return encryptedBytes
}
