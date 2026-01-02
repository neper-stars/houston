package crypto

import (
	"bytes"
	"testing"
)

func TestStarsRandomDeterministic(t *testing.T) {
	// Same seeds and rounds should produce the same sequence
	r1 := NewStarsRandom(17, 31, 5)
	r2 := NewStarsRandom(17, 31, 5)

	for i := 0; i < 100; i++ {
		v1 := r1.NextRandom()
		v2 := r2.NextRandom()
		if v1 != v2 {
			t.Errorf("iteration %d: random sequences diverged: %d != %d", i, v1, v2)
		}
	}
}

func TestStarsRandomDifferentSeeds(t *testing.T) {
	// Different seeds should produce different sequences
	r1 := NewStarsRandom(17, 31, 5)
	r2 := NewStarsRandom(19, 37, 5)

	// Check first few values are different
	allSame := true
	for i := 0; i < 10; i++ {
		if r1.NextRandom() != r2.NextRandom() {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("different seeds produced identical sequences")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		salt        int
		gameId      int
		turn        int
		playerIndex int
		shareware   int
	}{
		{
			name:        "simple data",
			data:        []byte("Hello, Stars!"),
			salt:        0x123,
			gameId:      42,
			turn:        1,
			playerIndex: 0,
			shareware:   0,
		},
		{
			name:        "aligned 4 bytes",
			data:        []byte("TEST"),
			salt:        0x456,
			gameId:      100,
			turn:        50,
			playerIndex: 3,
			shareware:   0,
		},
		{
			name:        "aligned 8 bytes",
			data:        []byte("TESTTEST"),
			salt:        0x789,
			gameId:      200,
			turn:        99,
			playerIndex: 7,
			shareware:   0,
		},
		{
			name:        "1 byte",
			data:        []byte{0x42},
			salt:        0x100,
			gameId:      1,
			turn:        1,
			playerIndex: 1,
			shareware:   0,
		},
		{
			name:        "2 bytes",
			data:        []byte{0x42, 0x43},
			salt:        0x200,
			gameId:      2,
			turn:        2,
			playerIndex: 2,
			shareware:   0,
		},
		{
			name:        "3 bytes",
			data:        []byte{0x42, 0x43, 0x44},
			salt:        0x300,
			gameId:      3,
			turn:        3,
			playerIndex: 3,
			shareware:   0,
		},
		{
			name:        "5 bytes (padding edge)",
			data:        []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			salt:        0x400,
			gameId:      4,
			turn:        4,
			playerIndex: 0,
			shareware:   0,
		},
		{
			name:        "with shareware flag",
			data:        []byte("shareware test"),
			salt:        0x500,
			gameId:      10,
			turn:        20,
			playerIndex: 1,
			shareware:   1,
		},
		{
			name:        "high salt bit 10 set",
			data:        []byte("high salt"),
			salt:        0x400, // bit 10 set
			gameId:      5,
			turn:        5,
			playerIndex: 5,
			shareware:   0,
		},
		{
			name:        "max salt value",
			data:        []byte("max salt test"),
			salt:        0x7FF, // all 11 bits set
			gameId:      255,
			turn:        255,
			playerIndex: 15,
			shareware:   1,
		},
		{
			name:        "binary data",
			data:        []byte{0x00, 0xFF, 0x55, 0xAA, 0x12, 0x34, 0x56, 0x78},
			salt:        0x1AB,
			gameId:      123,
			turn:        45,
			playerIndex: 6,
			shareware:   0,
		},
		{
			name:        "empty data",
			data:        []byte{},
			salt:        0x000,
			gameId:      0,
			turn:        0,
			playerIndex: 0,
			shareware:   0,
		},
		{
			name:        "larger data block",
			data:        bytes.Repeat([]byte{0xDE, 0xAD, 0xBE, 0xEF}, 100),
			salt:        0x2CD,
			gameId:      999,
			turn:        500,
			playerIndex: 8,
			shareware:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			enc := NewEncryptor()
			enc.InitEncryption(tt.salt, tt.gameId, tt.turn, tt.playerIndex, tt.shareware)
			encrypted := enc.EncryptBytes(tt.data)

			// Decrypt
			dec := NewDecryptor()
			dec.InitDecryption(tt.salt, tt.gameId, tt.turn, tt.playerIndex, tt.shareware)
			decrypted := dec.DecryptBytes(encrypted)

			// Verify round-trip
			if !bytes.Equal(decrypted, tt.data) {
				t.Errorf("round-trip failed:\noriginal:  %v\nencrypted: %v\ndecrypted: %v",
					tt.data, encrypted, decrypted)
			}

			// Verify encryption actually changed the data (unless empty)
			if len(tt.data) > 0 && bytes.Equal(encrypted, tt.data) {
				t.Error("encryption did not modify the data")
			}
		})
	}
}

func TestDecryptEncryptRoundTrip(t *testing.T) {
	// Test the reverse: decrypt then encrypt should also round-trip
	// (since XOR is symmetric)
	data := []byte("symmetric operation test")
	salt := 0x123
	gameId := 42
	turn := 10
	playerIndex := 2
	shareware := 0

	// Decrypt first (treating data as if it were encrypted)
	dec := NewDecryptor()
	dec.InitDecryption(salt, gameId, turn, playerIndex, shareware)
	decrypted := dec.DecryptBytes(data)

	// Encrypt back
	enc := NewEncryptor()
	enc.InitEncryption(salt, gameId, turn, playerIndex, shareware)
	reencrypted := enc.EncryptBytes(decrypted)

	if !bytes.Equal(reencrypted, data) {
		t.Errorf("decrypt-encrypt round-trip failed:\noriginal:    %v\ndecrypted:   %v\nreencrypted: %v",
			data, decrypted, reencrypted)
	}
}

func TestEncryptorDecryptorSymmetry(t *testing.T) {
	// Since XOR is symmetric, encrypting and decrypting should be identical operations
	data := []byte("XOR symmetry test data")
	salt := 0x2AB
	gameId := 77
	turn := 33
	playerIndex := 4
	shareware := 0

	enc := NewEncryptor()
	enc.InitEncryption(salt, gameId, turn, playerIndex, shareware)
	encrypted := enc.EncryptBytes(data)

	dec := NewDecryptor()
	dec.InitDecryption(salt, gameId, turn, playerIndex, shareware)
	decrypted := dec.DecryptBytes(data)

	// Both should produce the same output
	if !bytes.Equal(encrypted, decrypted) {
		t.Errorf("encryptor and decryptor produced different results:\nencrypted: %v\ndecrypted: %v",
			encrypted, decrypted)
	}
}

func TestPrimesTableLength(t *testing.T) {
	// The primes table should have exactly 64 entries
	if len(primes) != 64 {
		t.Errorf("expected 64 primes, got %d", len(primes))
	}
}

func TestPrimesTableIndexBounds(t *testing.T) {
	// Test that all possible salt values produce valid prime indices
	// Salt uses bits 0-4 for index1 and bits 5-9 for index2, with bit 10 switching halves
	for salt := 0; salt <= 0x7FF; salt++ {
		index1 := salt & 0x1F
		index2 := (salt >> 5) & 0x1F

		if (salt >> 10) == 1 {
			index1 += 32
		} else {
			index2 += 32
		}

		if index1 < 0 || index1 >= 64 {
			t.Errorf("salt %x: index1 %d out of bounds", salt, index1)
		}
		if index2 < 0 || index2 >= 64 {
			t.Errorf("salt %x: index2 %d out of bounds", salt, index2)
		}
	}
}

func TestDataLengthPreserved(t *testing.T) {
	// Encrypted/decrypted data should have the same length as input
	lengths := []int{0, 1, 2, 3, 4, 5, 7, 8, 9, 15, 16, 17, 100, 1000}

	for _, length := range lengths {
		data := make([]byte, length)
		for i := range data {
			data[i] = byte(i % 256)
		}

		enc := NewEncryptor()
		enc.InitEncryption(0x123, 1, 1, 1, 0)
		encrypted := enc.EncryptBytes(data)

		if len(encrypted) != len(data) {
			t.Errorf("length %d: encrypted length %d != original length %d",
				length, len(encrypted), len(data))
		}

		dec := NewDecryptor()
		dec.InitDecryption(0x123, 1, 1, 1, 0)
		decrypted := dec.DecryptBytes(encrypted)

		if len(decrypted) != len(data) {
			t.Errorf("length %d: decrypted length %d != original length %d",
				length, len(decrypted), len(data))
		}
	}
}
