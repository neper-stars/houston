package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/neper-stars/houston/blocks"
)

// Expected data structures for message scenario
type ExpectedMessage struct {
	Sender      string `json:"sender"`
	Recipient   string `json:"recipient"`
	Message     string `json:"message"`
	IsBroadcast bool   `json:"isBroadcast"`
}

type ExpectedMessageData struct {
	Scenario string            `json:"scenario"`
	Messages []ExpectedMessage `json:"messages"`
}

func loadMessageExpected(t *testing.T) *ExpectedMessageData {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-message", "expected.json")
	jsonData, err := os.ReadFile(path)
	require.NoError(t, err)

	var expected ExpectedMessageData
	err = json.Unmarshal(jsonData, &expected)
	require.NoError(t, err)
	return &expected
}

func loadMessageFile(t *testing.T, filename string) (FileData, []blocks.Block) {
	t.Helper()
	path := filepath.Join("..", "testdata", "scenario-message", filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err)
	return fd, blockList
}

func TestScenarioMessage_MessageBlocks(t *testing.T) {
	expected := loadMessageExpected(t)
	_, blockList := loadMessageFile(t, "game.m1")

	// Build player ID to name map
	playerNames := make(map[int]string)
	for _, block := range blockList {
		if p, ok := block.(blocks.PlayerBlock); ok {
			playerNames[p.PlayerNumber] = p.NamePlural
		}
	}

	// Get the receiving player's name (player 0 from the file header)
	var receivingPlayer string
	for _, block := range blockList {
		if fh, ok := block.(blocks.FileHeader); ok {
			receivingPlayer = playerNames[fh.PlayerIndex()]
			break
		}
	}

	// Collect messages
	var messages []blocks.MessageBlock
	for _, block := range blockList {
		if mb, ok := block.(blocks.MessageBlock); ok {
			messages = append(messages, mb)
		}
	}

	require.Equal(t, len(expected.Messages), len(messages),
		"Should have %d messages", len(expected.Messages))

	// Validate each message
	for i, exp := range expected.Messages {
		msg := messages[i]
		t.Run(exp.Sender, func(t *testing.T) {
			// Validate sender
			senderName := playerNames[msg.SenderId]
			assert.Equal(t, exp.Sender, senderName, "Sender should match")

			// Validate recipient
			if exp.IsBroadcast {
				assert.True(t, msg.IsBroadcast(), "Should be broadcast")
			} else {
				assert.False(t, msg.IsBroadcast(), "Should not be broadcast")
				// For direct messages, the recipient is the player whose M file we're reading
				assert.Equal(t, exp.Recipient, receivingPlayer, "Recipient should match")
			}

			// Validate message text
			assert.Equal(t, exp.Message, msg.Message, "Message text should match")

			// Validate broadcast flag
			assert.Equal(t, exp.IsBroadcast, msg.IsBroadcast(), "IsBroadcast should match")
		})
	}
}
