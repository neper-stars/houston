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

// Expected data structures for zip orders scenario
type ExpectedZipProdItem struct {
	ItemType int    `json:"itemType"`
	Quantity int    `json:"quantity"`
	Name     string `json:"name"`
}

type ExpectedZipProdDefault struct {
	NoResearch int                   `json:"fNoResearch"` // 0=contribute to research, 1=don't
	Items      []ExpectedZipProdItem `json:"items"`
}

type ExpectedZipOrdersPlayer struct {
	Number         int                    `json:"number"`
	RaceName       string                 `json:"raceName"`
	ZipProdDefault ExpectedZipProdDefault `json:"zipProdDefault"`
}

type ExpectedZipOrdersSaveAndSubmit struct {
	ZipProdDefault ExpectedZipProdDefault `json:"zipProdDefault"`
}

type ExpectedZipOrdersFile struct {
	Description   string                          `json:"description"`
	Player        *ExpectedZipOrdersPlayer        `json:"player,omitempty"`
	SaveAndSubmit *ExpectedZipOrdersSaveAndSubmit `json:"saveAndSubmit,omitempty"`
}

type ExpectedZipOrdersData struct {
	Scenario    string                           `json:"scenario"`
	Description string                           `json:"description"`
	Files       map[string]ExpectedZipOrdersFile `json:"files"`
}

// loadZipOrdersExpected loads the expected.json from a scenario directory
func loadZipOrdersExpected(t *testing.T, scenarioDir string) *ExpectedZipOrdersData {
	t.Helper()

	expectedPath := filepath.Join("..", "testdata", scenarioDir, "expected.json")

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Skipf("Test files not found in testdata/%s/", scenarioDir)
	}

	data, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "Failed to read expected.json")

	var expected ExpectedZipOrdersData
	err = json.Unmarshal(data, &expected)
	require.NoError(t, err, "Failed to parse expected.json")

	return &expected
}

// loadZipOrdersFile loads and parses a game file from a scenario directory
func loadZipOrdersFile(t *testing.T, scenarioDir, filename string) (FileData, []blocks.Block) {
	t.Helper()

	path := filepath.Join("..", "testdata", scenarioDir, filename)
	fileBytes, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to read %s", filename)

	fd := FileData(fileBytes)
	blockList, err := fd.BlockList()
	require.NoError(t, err, "BlockList() failed for %s", filename)

	return fd, blockList
}

// findPlayerBlock finds the first PlayerBlock in the block list
func findPlayerBlock(t *testing.T, blockList []blocks.Block) *blocks.PlayerBlock {
	t.Helper()

	for _, block := range blockList {
		if pb, ok := block.(blocks.PlayerBlock); ok {
			return &pb
		}
	}
	return nil
}

// verifyZipProdItems verifies zip prod items match expected
func verifyZipProdItems(t *testing.T, expected []ExpectedZipProdItem, actual []blocks.ZipProdQueueItem, context string) {
	t.Helper()

	require.Len(t, actual, len(expected), "%s: ZipProd item count should match", context)

	for i, expectedItem := range expected {
		actualItem := actual[i]
		assert.Equal(t, uint16(expectedItem.ItemType), actualItem.ItemType, //nolint:gosec // test data is controlled
			"%s: Item %d (%s) type should match", context, i, expectedItem.Name)
		assert.Equal(t, uint16(expectedItem.Quantity), actualItem.Quantity, //nolint:gosec // test data is controlled
			"%s: Item %d (%s) quantity should match", context, i, expectedItem.Name)
	}
}

func TestScenarioZipOrders_Year2(t *testing.T) {
	expected := loadZipOrdersExpected(t, "scenario-ziporders/year2")

	for filename, fileExpected := range expected.Files {
		t.Run(filename, func(t *testing.T) {
			_, blockList := loadZipOrdersFile(t, "scenario-ziporders/year2", filename)

			if fileExpected.Player != nil {
				player := findPlayerBlock(t, blockList)
				require.NotNil(t, player, "Player block should exist")

				assert.Equal(t, fileExpected.Player.Number, player.PlayerNumber, "Player number should match")
				assert.Equal(t, fileExpected.Player.RaceName, player.NameSingular, "Race name should match")
				expectedNoResearch := fileExpected.Player.ZipProdDefault.NoResearch != 0
				assert.Equal(t, expectedNoResearch, player.ZipProdDefault.NoResearch,
					"ZipProd fNoResearch should match")

				verifyZipProdItems(t, fileExpected.Player.ZipProdDefault.Items,
					player.ZipProdDefault.Items, "Player.ZipProdDefault")
			}
		})
	}
}

func TestScenarioZipOrders_Year2NewOrders(t *testing.T) {
	expected := loadZipOrdersExpected(t, "scenario-ziporders/year2-neworders")

	for filename, fileExpected := range expected.Files {
		t.Run(filename, func(t *testing.T) {
			_, blockList := loadZipOrdersFile(t, "scenario-ziporders/year2-neworders", filename)

			if fileExpected.Player != nil {
				player := findPlayerBlock(t, blockList)
				require.NotNil(t, player, "Player block should exist")

				assert.Equal(t, fileExpected.Player.Number, player.PlayerNumber, "Player number should match")
				assert.Equal(t, fileExpected.Player.RaceName, player.NameSingular, "Race name should match")
				expectedNoResearch := fileExpected.Player.ZipProdDefault.NoResearch != 0
				assert.Equal(t, expectedNoResearch, player.ZipProdDefault.NoResearch,
					"ZipProd fNoResearch should match")

				verifyZipProdItems(t, fileExpected.Player.ZipProdDefault.Items,
					player.ZipProdDefault.Items, "Player.ZipProdDefault")
			}

			if fileExpected.SaveAndSubmit != nil {
				// Find SaveAndSubmit block and verify zip prod data
				var saveAndSubmitData []byte
				for _, block := range blockList {
					if block.BlockTypeID() == blocks.SaveAndSubmitBlockType {
						saveAndSubmitData = block.DecryptedData()
						break
					}
				}
				require.NotNil(t, saveAndSubmitData, "SaveAndSubmit block should exist")

				// Parse zip prod from SaveAndSubmit block
				// Format: byte 0 = flags, byte 1 = item count, then 2 bytes per item
				require.GreaterOrEqual(t, len(saveAndSubmitData), 2, "SaveAndSubmit should have at least 2 bytes")

				fNoResearch := saveAndSubmitData[0]
				itemCount := int(saveAndSubmitData[1])

				expectedNoResearch := byte(0)
				if fileExpected.SaveAndSubmit.ZipProdDefault.NoResearch != 0 {
					expectedNoResearch = 1
				}
				assert.Equal(t, expectedNoResearch, fNoResearch,
					"SaveAndSubmit ZipProd fNoResearch should match")
				assert.Equal(t, len(fileExpected.SaveAndSubmit.ZipProdDefault.Items), itemCount,
					"SaveAndSubmit ZipProd item count should match")

				// Parse and verify items
				var actualItems []blocks.ZipProdQueueItem
				for i := 0; i < itemCount && 2+i*2+1 < len(saveAndSubmitData); i++ {
					val := uint16(saveAndSubmitData[2+i*2]) | uint16(saveAndSubmitData[2+i*2+1])<<8
					actualItems = append(actualItems, blocks.ZipProdQueueItem{
						ItemType: val & 0x3F,
						Quantity: val >> 6,
					})
				}

				verifyZipProdItems(t, fileExpected.SaveAndSubmit.ZipProdDefault.Items,
					actualItems, "SaveAndSubmit.ZipProdDefault")
			}
		})
	}
}
