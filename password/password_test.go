package password

import (
	"encoding/binary"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func read32(bytes []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(bytes[offset:])
}

func TestHashAndGuess(t *testing.T) {
	expectedHash := uint32(156085230)
	// The AI password and alternatives that show the weakness in the hash:
	h1 := HashRacePassword("viewai")  // this is the "real" password
	h2 := HashRacePassword("fymmgsd") // this is a valid collision
	h3 := HashRacePassword("yfmmgsd") // ditto
	h4 := HashRacePassword("iymtfi")  // ditto

	// all those "unique" password have the same hash... surprise :p
	assert.Equal(t, expectedHash, h1)
	assert.Equal(t, expectedHash, h2)
	assert.Equal(t, expectedHash, h3)
	assert.Equal(t, expectedHash, h4)

	// This shows golang's implicit integer overflow protection by returning a long integer :-/
	h := HashRacePassword("aaaaaaaa")
	assert.Equal(t, uint32(86857028), h)

	// Brute force the "aaba" password, note how to gain speed I restrict the charset
	// to only a and b. In a real life scenario you'll want to at least add all the ascii
	// characters
	matches := GuessRacePassword(read32([]byte{67, 18, 14, 0}, 0), 4, 1, "ab", false)
	require.Equal(t, len(matches), 1)
	assert.Equal(t, "aaba", matches[0])

	realPass := "aaaa"
	h = HashRacePassword(realPass)
	matches = GuessRacePassword(h, 4, 1, "a", false)

	require.Equal(t, len(matches), 1)
	assert.Equal(t, realPass, matches[0])

	// a more intensive computation
	realPass = "azert"
	aToZ := "abcdefghijklmnopqrstuvwxyz"
	h = HashRacePassword(realPass)
	matches = GuessRacePassword(h, 5, 1, aToZ, false)
	// fmt.Println(matches)
	assert.Equal(t, realPass, matches[0])

	// real life observed data
	h = HashRacePassword("flubu")
	byteArray := make([]byte, 4)
	// binary encode in little endian format
	binary.LittleEndian.PutUint32(byteArray, h)
	assert.Equal(t, []byte{79, 166, 16, 0}, byteArray)

	// the charset is reduced on purpose to speed up the search.
	// the purpose of this test is to validate that our guessing function
	// finds a proper password from an already encoded hash extracted from a real
	// game file. This avoids the auto validation test flaw that only proves
	// our guess function is able to guess our hash algo and not a real live data.
	matches = GuessRacePassword(read32([]byte{79, 166, 16, 0}, 0), 5, 10, "abflu", false)
	require.True(t, 1 < len(matches))
	assert.Equal(t, "flubu", matches[0])
}

func TestHashRacePasswordBytes(t *testing.T) {
	// Verify bytes version matches string version
	testCases := []string{"viewai", "test", "aaaaaaaa", "flubu"}
	for _, tc := range testCases {
		expected := HashRacePassword(tc)
		actual := HashRacePasswordBytes([]byte(tc))
		assert.Equal(t, expected, actual, "mismatch for %q", tc)
	}
}

func TestGuessRacePasswordParallel(t *testing.T) {
	// Test basic functionality
	realPass := "aaba"
	h := HashRacePassword(realPass)
	matches := GuessRacePasswordParallel(h, 4, 1, "ab", 0, nil)
	require.Equal(t, 1, len(matches))
	assert.Equal(t, realPass, matches[0])

	// Test with single character charset
	realPass = "aaaa"
	h = HashRacePassword(realPass)
	matches = GuessRacePasswordParallel(h, 4, 1, "a", 0, nil)
	require.Equal(t, 1, len(matches))
	assert.Equal(t, realPass, matches[0])

	// Test with larger charset
	realPass = "azert"
	aToZ := "abcdefghijklmnopqrstuvwxyz"
	h = HashRacePassword(realPass)
	matches = GuessRacePasswordParallel(h, 5, 1, aToZ, 0, nil)
	require.GreaterOrEqual(t, len(matches), 1)
	// The first match might not be the original due to hash collisions,
	// but it should hash to the same value
	assert.Equal(t, h, HashRacePassword(matches[0]))

	// Test multiple matches
	matches = GuessRacePasswordParallel(read32([]byte{79, 166, 16, 0}, 0), 5, 10, "abflu", 0, nil)
	require.True(t, len(matches) >= 1)
	// All matches should have the same hash
	targetHash := read32([]byte{79, 166, 16, 0}, 0)
	for _, m := range matches {
		assert.Equal(t, targetHash, HashRacePassword(m))
	}
}

func TestGuessRacePasswordParallelWithProgress(t *testing.T) {
	var progressCalled bool
	var lastCount uint64

	progress := func(tried uint64) {
		progressCalled = true
		lastCount = tried
	}

	realPass := "test"
	h := HashRacePassword(realPass)
	matches := GuessRacePasswordParallel(h, 4, 1, "abcdefghijklmnopqrstuvwxyz", 0, progress)

	require.GreaterOrEqual(t, len(matches), 1)
	assert.True(t, progressCalled, "progress callback should have been called")
	assert.Greater(t, lastCount, uint64(0), "should have tried some passwords")
}

func TestGuessRacePasswordParallelRealFile(t *testing.T) {
	// Test against a real race file with known password "f00ls"
	// The hash is extracted from the PlayerBlock in the race file
	expectedHash := uint32(534067)
	knownPassword := "f00ls"

	// Verify the known password hashes correctly
	assert.Equal(t, expectedHash, HashRacePassword(knownPassword))

	// Find passwords using the parallel guesser
	charset := "abcdefghijklmnopqrstuvwxyz0123456789"
	matches := GuessRacePasswordParallel(expectedHash, 5, 5, charset, 0, nil)

	require.GreaterOrEqual(t, len(matches), 1, "should find at least one match")

	// All matches should hash to the expected value
	for _, m := range matches {
		assert.Equal(t, expectedHash, HashRacePassword(m), "match %q should hash to %d", m, expectedHash)
	}

	// The known password should be among the matches (if we find enough)
	foundKnown := false
	for _, m := range matches {
		if m == knownPassword {
			foundKnown = true
			break
		}
	}
	// Note: Due to hash collisions, the known password might not be in the first 5 matches
	// but we verify all found matches are valid
	t.Logf("Found %d matches, known password in results: %v", len(matches), foundKnown)
	t.Logf("Matches: %v", matches)
}

// Benchmarks to compare sequential vs parallel performance
func BenchmarkGuessRacePassword(b *testing.B) {
	h := HashRacePassword("test")
	charset := "abcdefghijklmnopqrstuvwxyz"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GuessRacePassword(h, 4, 1, charset, false)
	}
}

func BenchmarkGuessRacePasswordParallel(b *testing.B) {
	h := HashRacePassword("test")
	charset := "abcdefghijklmnopqrstuvwxyz"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GuessRacePasswordParallel(h, 4, 1, charset, 0, nil)
	}
}

func BenchmarkGuessRacePasswordParallel1Worker(b *testing.B) {
	h := HashRacePassword("test")
	charset := "abcdefghijklmnopqrstuvwxyz"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GuessRacePasswordParallel(h, 4, 1, charset, 1, nil)
	}
}

func BenchmarkGuessRacePasswordLonger(b *testing.B) {
	h := HashRacePassword("azert")
	charset := "abcdefghijklmnopqrstuvwxyz"

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GuessRacePassword(h, 5, 1, charset, false)
		}
	})

	b.Run("Parallel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GuessRacePasswordParallel(h, 5, 1, charset, 0, nil)
		}
	})

	b.Run("Parallel-"+string(rune('0'+runtime.NumCPU()))+"cores", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GuessRacePasswordParallel(h, 5, 1, charset, runtime.NumCPU(), nil)
		}
	})
}
