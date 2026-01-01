package password

import (
	"runtime"
	"sync"
	"sync/atomic"

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
	return HashRacePasswordBytes([]byte(inputString))
}

// HashRacePasswordBytes computes the weak hash of a race password from bytes.
// This avoids string allocation in hot paths.
func HashRacePasswordBytes(charList []byte) uint32 {
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

// ProgressCallback is called periodically with the number of passwords tried.
type ProgressCallback func(tried uint64)

// GuessRacePasswordParallel guesses a race file's password using parallel workers.
// This is significantly faster than GuessRacePassword on multi-core systems.
//
// Parameters:
//   - hash: the target hash to find collisions for
//   - maxLength: maximum password length to try
//   - matchesAllowed: stop after finding this many matches (0 = unlimited)
//   - charset: characters to use for brute force
//   - workers: number of parallel workers (0 = use all CPU cores)
//   - progress: optional callback for progress reporting (can be nil)
//
// Returns a slice of matching passwords.
func GuessRacePasswordParallel(hash uint32, maxLength, matchesAllowed int,
	charset string, workers int, progress ProgressCallback) []string {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	charsetBytes := []byte(charset)
	charsetLen := len(charsetBytes)

	if charsetLen == 0 || maxLength == 0 {
		return nil
	}

	// Channel for results
	resultCh := make(chan string, 100)

	// Atomic counters
	var matchCount atomic.Int64
	var triedCount atomic.Uint64
	var done atomic.Bool

	// WaitGroup for workers
	var wg sync.WaitGroup

	// Partition work by first character
	// Each worker handles passwords starting with specific first characters
	charsPerWorker := (charsetLen + workers - 1) / workers

	for w := 0; w < workers; w++ {
		startIdx := w * charsPerWorker
		endIdx := startIdx + charsPerWorker
		if endIdx > charsetLen {
			endIdx = charsetLen
		}
		if startIdx >= charsetLen {
			break
		}

		wg.Add(1)
		go func(firstCharStart, firstCharEnd int) {
			defer wg.Done()

			// Local buffer for password building (avoids allocations)
			buf := make([]byte, maxLength)

			// Process each first character assigned to this worker
			for firstIdx := firstCharStart; firstIdx < firstCharEnd; firstIdx++ {
				if done.Load() {
					return
				}

				buf[0] = charsetBytes[firstIdx]

				// Generate all combinations with this first character
				workerGenerate(buf, 1, maxLength, charsetBytes, hash,
					matchesAllowed, &matchCount, &triedCount, &done, resultCh)
			}
		}(startIdx, endIdx)
	}

	// Progress reporter goroutine
	var progressWg sync.WaitGroup
	if progress != nil {
		progressWg.Add(1)
		go func() {
			defer progressWg.Done()
			for !done.Load() {
				progress(triedCount.Load())
				runtime.Gosched()
			}
		}()
	}

	// Collector goroutine
	var matches []string
	var collectorWg sync.WaitGroup
	collectorWg.Add(1)
	go func() {
		defer collectorWg.Done()
		for match := range resultCh {
			matches = append(matches, match)
		}
	}()

	// Wait for all workers to finish
	wg.Wait()
	done.Store(true)
	close(resultCh)

	// Wait for collector
	collectorWg.Wait()
	progressWg.Wait()

	// Final progress update
	if progress != nil {
		progress(triedCount.Load())
	}

	return matches
}

// workerGenerate recursively generates and tests password combinations.
// It uses a pre-allocated buffer to avoid string allocations.
func workerGenerate(buf []byte, pos, maxLen int, charset []byte, targetHash uint32,
	matchesAllowed int, matchCount *atomic.Int64, triedCount *atomic.Uint64,
	done *atomic.Bool, resultCh chan<- string) {
	// Test current password (length = pos)
	if pos > 0 {
		password := buf[:pos]
		triedCount.Add(1)

		if HashRacePasswordBytes(password) == targetHash {
			// Found a match
			count := matchCount.Add(1)
			resultCh <- string(password)

			if matchesAllowed > 0 && int(count) >= matchesAllowed {
				done.Store(true)
				return
			}
		}
	}

	// Generate longer passwords
	if pos >= maxLen || done.Load() {
		return
	}

	for _, c := range charset {
		if done.Load() {
			return
		}
		buf[pos] = c
		workerGenerate(buf, pos+1, maxLen, charset, targetHash,
			matchesAllowed, matchCount, triedCount, done, resultCh)
	}
}

func init() {
	for i := 0; i < 128; i++ {
		AsciiString += string(rune(i))
	}
}
