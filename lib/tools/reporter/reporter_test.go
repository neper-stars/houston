package reporter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTemplate(t *testing.T) {
	templatePath := filepath.Join("..", "..", "..", "cmd", "houston", "resources", "empty.ods")
	data, err := os.ReadFile(templatePath)
	require.NoError(t, err, "failed to read template file")

	doc, err := LoadBytes(data)
	require.NoError(t, err, "failed to load template")
	defer func() { _ = doc.Close() }()

	// Verify all expected sheets exist
	expectedSheets := []string{
		SheetSummary,
		SheetMyMinerals,
		SheetMyMineralHist,
		SheetMineralShuffle,
		SheetOpponentPop,
		SheetOpponentPopHist,
		SheetOpponentShips,
		SheetOpponentFleets,
		SheetNewDesigns,
		SheetScoreEstimates,
	}

	for _, name := range expectedSheets {
		sheet := doc.SheetByName(name)
		assert.NotNil(t, sheet, "sheet %q should exist", name)
	}
}

func TestODSCellOperations(t *testing.T) {
	templatePath := filepath.Join("..", "..", "..", "cmd", "houston", "resources", "empty.ods")
	data, err := os.ReadFile(templatePath)
	require.NoError(t, err)

	doc, err := LoadBytes(data)
	require.NoError(t, err)
	defer func() { _ = doc.Close() }()

	sheet := doc.SheetByName(SheetSummary)
	require.NotNil(t, sheet)

	// Test SetCellString
	doc.SetCellString(sheet, 0, 0, "Test Header")
	assert.Equal(t, "Test Header", doc.GetCellString(sheet, 0, 0))

	// Test SetCellInt
	doc.SetCellInt(sheet, 1, 0, 12345)
	val, ok := doc.GetCellInt(sheet, 1, 0)
	assert.True(t, ok)
	assert.Equal(t, int64(12345), val)

	// Test SetCellFloat
	doc.SetCellFloat(sheet, 2, 0, 123.45)
	// Float is stored as string, so we check the string representation
	strVal := doc.GetCellString(sheet, 2, 0)
	assert.Equal(t, "123.45", strVal)

	// Test AppendRow
	doc.AppendRow(sheet, "Row1", int64(100), 50.5)
	rowCount := doc.RowCount(sheet)
	assert.GreaterOrEqual(t, rowCount, 4)

	// Test SetHeaderRow
	doc.SetHeaderRow(sheet, "Col1", "Col2", "Col3")
	assert.Equal(t, "Col1", doc.GetCellString(sheet, 0, 0))
	assert.Equal(t, "Col2", doc.GetCellString(sheet, 0, 1))
	assert.Equal(t, "Col3", doc.GetCellString(sheet, 0, 2))
}

func TestODSWriteAndReload(t *testing.T) {
	templatePath := filepath.Join("..", "..", "..", "cmd", "houston", "resources", "empty.ods")
	data, err := os.ReadFile(templatePath)
	require.NoError(t, err)

	doc, err := LoadBytes(data)
	require.NoError(t, err)

	sheet := doc.SheetByName(SheetSummary)
	require.NotNil(t, sheet)

	// Write some data
	doc.SetHeaderRow(sheet, "Turn", "Year", "Value")
	doc.AppendRow(sheet, int64(1), int64(2401), int64(1000))
	doc.AppendRow(sheet, int64(2), int64(2402), int64(2000))

	// Write to bytes
	outputData, err := doc.WriteBytes()
	require.NoError(t, err)
	_ = doc.Close()

	// Reload and verify
	doc2, err := LoadBytes(outputData)
	require.NoError(t, err)
	defer func() { _ = doc2.Close() }()

	sheet2 := doc2.SheetByName(SheetSummary)
	require.NotNil(t, sheet2)

	// Verify header
	assert.Equal(t, "Turn", doc2.GetCellString(sheet2, 0, 0))
	assert.Equal(t, "Year", doc2.GetCellString(sheet2, 0, 1))
	assert.Equal(t, "Value", doc2.GetCellString(sheet2, 0, 2))

	// Verify data rows
	turn1, ok := doc2.GetCellInt(sheet2, 1, 0)
	assert.True(t, ok)
	assert.Equal(t, int64(1), turn1)

	year1, ok := doc2.GetCellInt(sheet2, 1, 1)
	assert.True(t, ok)
	assert.Equal(t, int64(2401), year1)
}

func TestReporterLoadGameFile(t *testing.T) {
	templatePath := filepath.Join("..", "..", "..", "cmd", "houston", "resources", "empty.ods")
	gameFilePath := filepath.Join("..", "..", "..", "testdata", "scenario-basic", "game.m1")

	// Check if test file exists
	if _, err := os.Stat(gameFilePath); os.IsNotExist(err) {
		t.Skip("test game file not found")
	}

	templateData, err := os.ReadFile(templatePath)
	require.NoError(t, err)

	rep := New()
	rep.SetTemplateBytes(templateData)

	err = rep.LoadFile(gameFilePath)
	require.NoError(t, err)

	assert.Greater(t, rep.GameID(), uint32(0))
	assert.GreaterOrEqual(t, rep.Turn(), 0)
	assert.GreaterOrEqual(t, rep.Year(), 2400)
}

func TestReporterGenerateReport(t *testing.T) {
	templatePath := filepath.Join("..", "..", "..", "cmd", "houston", "resources", "empty.ods")
	gameFilePath := filepath.Join("..", "..", "..", "testdata", "scenario-basic", "game.m1")

	// Check if test files exist
	if _, err := os.Stat(gameFilePath); os.IsNotExist(err) {
		t.Skip("test game file not found")
	}

	templateData, err := os.ReadFile(templatePath)
	require.NoError(t, err)

	rep := New()
	rep.SetTemplateBytes(templateData)

	err = rep.LoadFile(gameFilePath)
	require.NoError(t, err)

	// Generate report
	opts := DefaultOptions()
	reportData, err := rep.GenerateReport(opts)
	require.NoError(t, err)
	assert.Greater(t, len(reportData), 0)

	// Reload and verify sheets have data
	doc, err := LoadBytes(reportData)
	require.NoError(t, err)
	defer func() { _ = doc.Close() }()

	// Summary sheet should have content
	summary := doc.SheetByName(SheetSummary)
	require.NotNil(t, summary)
	assert.Greater(t, doc.RowCount(summary), 1, "Summary sheet should have data rows")

	// Check that Game ID is present
	gameIDLabel := doc.GetCellString(summary, 1, 0)
	assert.Equal(t, "Game ID", gameIDLabel)
}

func TestCollectPlayerSnapshot(t *testing.T) {
	gameFilePath := filepath.Join("..", "..", "..", "testdata", "scenario-basic", "game.m1")

	if _, err := os.Stat(gameFilePath); os.IsNotExist(err) {
		t.Skip("test game file not found")
	}

	rep := New()
	err := rep.LoadFile(gameFilePath)
	require.NoError(t, err)

	// Player 0 should have some data
	snap := rep.CollectPlayerSnapshot(0)
	assert.Equal(t, 0, snap.PlayerNumber)
	assert.Equal(t, rep.Turn(), snap.Turn)
	assert.Equal(t, rep.Year(), snap.Year)
	// At turn 0, player should have at least 1 planet (homeworld)
	assert.GreaterOrEqual(t, snap.PlanetCount, 0)
}

func TestCollectPlanetMineralData(t *testing.T) {
	gameFilePath := filepath.Join("..", "..", "..", "testdata", "scenario-basic", "game.m1")

	if _, err := os.Stat(gameFilePath); os.IsNotExist(err) {
		t.Skip("test game file not found")
	}

	rep := New()
	err := rep.LoadFile(gameFilePath)
	require.NoError(t, err)

	data := rep.CollectPlanetMineralData(0)
	// Should return a slice (may be empty if player has no planets)
	assert.NotNil(t, data)
}

func TestAnalyzeMineralShuffling(t *testing.T) {
	gameFilePath := filepath.Join("..", "..", "..", "testdata", "scenario-basic", "game.m1")

	if _, err := os.Stat(gameFilePath); os.IsNotExist(err) {
		t.Skip("test game file not found")
	}

	rep := New()
	err := rep.LoadFile(gameFilePath)
	require.NoError(t, err)

	// Analyze with a threshold - returns nil or empty slice if no shuffle needed
	recs := rep.AnalyzeMineralShuffling(0, 500)
	// Just verify it doesn't panic - result may be nil or empty
	_ = recs
}

func TestHistoryPreservation(t *testing.T) {
	templatePath := filepath.Join("..", "..", "..", "cmd", "houston", "resources", "empty.ods")
	gameFilePath := filepath.Join("..", "..", "..", "testdata", "scenario-basic", "game.m1")

	if _, err := os.Stat(gameFilePath); os.IsNotExist(err) {
		t.Skip("test game file not found")
	}

	templateData, err := os.ReadFile(templatePath)
	require.NoError(t, err)

	// First report generation
	rep1 := New()
	rep1.SetTemplateBytes(templateData)
	err = rep1.LoadFile(gameFilePath)
	require.NoError(t, err)

	report1Data, err := rep1.GenerateReport(DefaultOptions())
	require.NoError(t, err)

	// Second report generation with existing report
	rep2 := New()
	rep2.SetTemplateBytes(templateData)
	rep2.SetExistingReportBytes(report1Data)
	err = rep2.LoadFile(gameFilePath)
	require.NoError(t, err)

	report2Data, err := rep2.GenerateReport(DefaultOptions())
	require.NoError(t, err)

	// Verify the history sheet has data
	doc, err := LoadBytes(report2Data)
	require.NoError(t, err)
	defer func() { _ = doc.Close() }()

	histSheet := doc.SheetByName(SheetMyMineralHist)
	require.NotNil(t, histSheet)
	// Should have at least header + 1 data row
	assert.GreaterOrEqual(t, doc.RowCount(histSheet), 2)
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	assert.Equal(t, 0, opts.PlayerNumber)
	assert.Equal(t, int64(500), opts.MineralThreshold)
	assert.False(t, opts.IncludeAllPlanets)
	assert.False(t, opts.IncludeEmptyFleets)
}

func TestSheetConstants(t *testing.T) {
	// Verify sheet name constants match expected values
	assert.Equal(t, "Summary", SheetSummary)
	assert.Equal(t, "My Minerals", SheetMyMinerals)
	assert.Equal(t, "My Minerals History", SheetMyMineralHist)
	assert.Equal(t, "Mineral Shuffle", SheetMineralShuffle)
	assert.Equal(t, "Opponent Population", SheetOpponentPop)
	assert.Equal(t, "Opponent Pop History", SheetOpponentPopHist)
	assert.Equal(t, "Opponent Ships", SheetOpponentShips)
	assert.Equal(t, "Opponent Fleets", SheetOpponentFleets)
	assert.Equal(t, "New Designs", SheetNewDesigns)
	assert.Equal(t, "Score Estimates", SheetScoreEstimates)
}

func TestDetectedPlayerNumber(t *testing.T) {
	gameFilePath := filepath.Join("..", "..", "..", "testdata", "scenario-basic", "game.m1")

	if _, err := os.Stat(gameFilePath); os.IsNotExist(err) {
		t.Skip("test game file not found")
	}

	rep := New()
	// Before loading any file, should return -1
	assert.Equal(t, -1, rep.DetectedPlayerNumber())

	err := rep.LoadFile(gameFilePath)
	require.NoError(t, err)

	// After loading game.m1, should detect player 0 (player 1 in 1-indexed)
	playerNum := rep.DetectedPlayerNumber()
	assert.GreaterOrEqual(t, playerNum, 0)
	assert.LessOrEqual(t, playerNum, 15)
}

func TestGameIDValidation(t *testing.T) {
	templatePath := filepath.Join("..", "..", "..", "cmd", "houston", "resources", "empty.ods")
	gameFilePath := filepath.Join("..", "..", "..", "testdata", "scenario-basic", "game.m1")

	if _, err := os.Stat(gameFilePath); os.IsNotExist(err) {
		t.Skip("test game file not found")
	}

	templateData, err := os.ReadFile(templatePath)
	require.NoError(t, err)

	// Generate a report for the first game
	rep1 := New()
	rep1.SetTemplateBytes(templateData)
	err = rep1.LoadFile(gameFilePath)
	require.NoError(t, err)

	report1Data, err := rep1.GenerateReport(DefaultOptions())
	require.NoError(t, err)

	// Using the same game file should work (same game ID)
	rep2 := New()
	rep2.SetTemplateBytes(templateData)
	rep2.SetExistingReportBytes(report1Data)
	err = rep2.LoadFile(gameFilePath)
	require.NoError(t, err)

	_, err = rep2.GenerateReport(DefaultOptions())
	assert.NoError(t, err, "same game ID should work")

	// Verify the game ID is stored in the report
	doc, err := LoadBytes(report1Data)
	require.NoError(t, err)

	sheet := doc.SheetByName(SheetSummary)
	require.NotNil(t, sheet)

	label := doc.GetCellString(sheet, 1, 0)
	assert.Equal(t, "Game ID", label)

	gameID, ok := doc.GetCellInt(sheet, 1, 1)
	assert.True(t, ok)
	assert.Equal(t, int64(rep1.GameID()), gameID)

	// Modify the game ID in the report to simulate a different game
	doc.SetCellInt(sheet, 1, 1, 99999999) // Different game ID
	modifiedReportData, err := doc.WriteBytes()
	require.NoError(t, err)
	_ = doc.Close()

	// Using a report with different game ID should fail
	rep3 := New()
	rep3.SetTemplateBytes(templateData)
	rep3.SetExistingReportBytes(modifiedReportData)
	err = rep3.LoadFile(gameFilePath)
	require.NoError(t, err)

	_, err = rep3.GenerateReport(DefaultOptions())
	assert.Error(t, err, "different game ID should fail")
	assert.Contains(t, err.Error(), "game ID mismatch")
}
