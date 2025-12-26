package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/crypto"
	"github.com/neper-stars/houston/parser"
)

func TestComputeRaceFooter(t *testing.T) {
	// Collect all race files from testdata/scenario-racefiles
	var raceFiles []string

	// Files directly in scenario-racefiles
	directFiles, err := filepath.Glob("../testdata/scenario-racefiles/*.r*")
	if err != nil {
		t.Fatalf("Failed to glob direct race files: %v", err)
	}
	raceFiles = append(raceFiles, directFiles...)

	// Files in subdirectories
	subDirFiles, err := filepath.Glob("../testdata/scenario-racefiles/**/*.r*")
	if err != nil {
		t.Fatalf("Failed to glob subdirectory race files: %v", err)
	}
	raceFiles = append(raceFiles, subDirFiles...)

	if len(raceFiles) == 0 {
		t.Fatal("No race files found in testdata/scenario-racefiles")
	}

	t.Logf("Testing %d race files", len(raceFiles))

	for _, file := range raceFiles {
		t.Run(filepath.Base(file), func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", file, err)
			}

			fd := parser.FileData(data)
			blockList, err := fd.BlockList()
			if err != nil {
				t.Fatalf("Failed to parse %s: %v", file, err)
			}

			// Extract header, player block, and footer
			var salt int
			var encryptedData []byte
			var actualFooter []byte
			var singularName, pluralName string

			for _, blk := range blockList {
				switch b := blk.(type) {
				case blocks.FileHeader:
					salt = b.Salt()
				case *blocks.FileHeader:
					salt = b.Salt()
				}
				if blk.BlockTypeID() == blocks.PlayerBlockType {
					encryptedData = blk.BlockData()
					// Get the parsed race names from the PlayerBlock
					// Parser returns PlayerBlock by value, not pointer
					if pb, ok := blk.(blocks.PlayerBlock); ok {
						singularName = pb.NameSingular
						pluralName = pb.NamePlural
					}
				} else if blk.BlockTypeID() == blocks.FileFooterBlockType {
					actualFooter = blk.BlockData()
				}
			}

			if len(encryptedData) == 0 {
				t.Fatalf("No PlayerBlock found in %s", file)
			}
			if len(actualFooter) < 2 {
				t.Fatalf("No valid footer found in %s", file)
			}

			// Decrypt the PlayerBlock (race files: gameId=0, turn=0, playerIndex=31)
			decryptor := crypto.NewDecryptor()
			decryptor.InitDecryption(salt, 0, 0, 31, 0)
			decryptedData := decryptor.DecryptBytes(encryptedData)

			// Actual footer as 16-bit value
			actualFooterVal := uint16(actualFooter[0]) | uint16(actualFooter[1])<<8

			// Compute expected footer using race names
			computedFooter := ComputeRaceFooter(decryptedData, singularName, pluralName)

			if computedFooter != actualFooterVal {
				t.Errorf("Footer mismatch for %s: computed=0x%04X, actual=0x%04X (singular=%q, plural=%q)",
					file, computedFooter, actualFooterVal, singularName, pluralName)
			}
		})
	}
}
