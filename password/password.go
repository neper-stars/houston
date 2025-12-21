package password

import (
	"github.com/neper-stars/houston/log"
)

// AsciiString contains all ASCII characters for password guessing
var AsciiString string

func generateCombinations(charset string, maxLength int) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		generate("", charset, maxLength, ch)
	}()

	return ch
}

func generate(prefix string, charset string, maxLength int, ch chan<- string) {
	if maxLength == 0 {
		return
	}

	for _, char := range charset {
		newCombination := prefix + string(char)
		ch <- newCombination
		generate(newCombination, charset, maxLength-1, ch)
	}
}

// HashRacePassword computes the weak hash of a race password
func HashRacePassword(inputString string) uint32 {
	charList := []byte(inputString)

	// Start with first char value
	output := uint32(charList[0])

	// Now perform an arithmetic operation on each of the following char values
	// We'll simulate 32-bit integer overflow by taking python's long type and
	// apply a bit mask of 0xffffffff
	for index, char := range charList[1:] {
		if index%2 == 0 {
			// if the character index is odd, multiply it against our hash
			output = (output * uint32(char)) & 0xffffffff
		} else {
			// else add it to our hash
			output = (output + uint32(char)) & 0xffffffff
		}
	}

	return output
}

// GuessRacePassword guesses a race file's password by brute force.
//
// NOTE: you will burn out your CPU if you have too liberal of settings here
//
// Because the hashing algorithm is incredibly weak, this will most likely find
// a number of alternative strings to use instead of the original password
//
// Used like so:
// GuessRacePassword(hash, maxLength, matchesAllowed, charset, verbose)
//
// hash is the value found in the 4 byte hash generated in a PlayerBlock
// at data offset 12
//
// maxLength = the length of characters in the password.
//
// matchesAllowed = quit after this many matches are found
//
// charset = array of ascii values to produce combinations from
func GuessRacePassword(hash uint32, maxLength int, matchesAllowed int, charset string, verbose bool) []string {
	var matches []string

	for combination := range generateCombinations(charset, maxLength) {
		possibleHash := HashRacePassword(combination)
		if possibleHash == hash {
			matches = append(matches, combination)
			if verbose {
				log.Debug("found password", log.F("password", combination))
			}
		}
		if len(matches) == matchesAllowed {
			return matches
		}
	}

	return matches
}

func init() {
	for i := 0; i < 128; i++ {
		AsciiString += string(rune(i))
	}
}
