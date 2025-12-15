package houston

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
