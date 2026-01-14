package reporter

import (
	"fmt"
	"sort"

	"github.com/neper-stars/houston/data"
)

// generateSummarySheet creates the Summary sheet.
func (r *Reporter) generateSummarySheet(doc *ODSDocument, opts *ReportOptions) error {
	sheet := doc.SheetByName(SheetSummary)
	if sheet == nil {
		return fmt.Errorf("sheet %q not found", SheetSummary)
	}

	// Clear and set headers
	doc.ClearSheet(sheet, 0)
	doc.SetHeaderRow(sheet, "Game Info", "Value")

	// Game info
	doc.AppendRow(sheet, "Game ID", int64(r.GameID()))
	doc.AppendRow(sheet, "Turn", int64(r.Turn()))
	doc.AppendRow(sheet, "Year", int64(r.Year()))
	doc.AppendRow(sheet, "", "")

	// Player summary header
	doc.AppendRow(sheet, "Player", "Planets", "Pop (M)", "Ironium", "Boranium", "Germanium", "Ships", "Est Score")

	// Add all players we have visibility of
	for _, player := range r.store.AllPlayers() {
		// Skip opponents we have no visibility data on
		if player.PlayerNumber != opts.PlayerNumber && !r.HasVisibilityOf(player.PlayerNumber) {
			continue
		}

		snap := r.CollectPlayerSnapshot(player.PlayerNumber)

		doc.AppendRow(sheet,
			player.NamePlural,
			int64(snap.PlanetCount),
			snap.TotalPopulation/1000000, // Convert to millions
			snap.Minerals.Ironium,
			snap.Minerals.Boranium,
			snap.Minerals.Germanium,
			int64(snap.Ships.Total),
			int64(snap.EstimatedScore),
		)
	}

	return nil
}

// generateMyMineralsSheet creates the My Minerals sheet.
func (r *Reporter) generateMyMineralsSheet(doc *ODSDocument, opts *ReportOptions) error {
	sheet := doc.SheetByName(SheetMyMinerals)
	if sheet == nil {
		return fmt.Errorf("sheet %q not found", SheetMyMinerals)
	}

	// Clear and set headers
	doc.ClearSheet(sheet, 0)
	doc.SetHeaderRow(sheet, "Planet", "Ironium", "Boranium", "Germanium", "En Route I", "En Route B", "En Route G", "Total I", "Total B", "Total G", "Population")

	// Get planet data
	planetData := r.CollectPlanetMineralData(opts.PlayerNumber)

	// Track data rows for formula range (row 2 onwards, 1-indexed in formulas)
	dataStartRow := 2
	dataRowCount := 0

	for _, p := range planetData {
		if !opts.IncludeAllPlanets && p.Ironium == 0 && p.Boranium == 0 && p.Germanium == 0 {
			continue
		}

		currentRow := dataStartRow + dataRowCount // 1-indexed row number for formulas
		dataRowCount++

		// Columns: A=Planet, B=Ironium, C=Boranium, D=Germanium, E=EnRouteI, F=EnRouteB, G=EnRouteG, H=TotalI, I=TotalB, J=TotalG, K=Pop
		doc.AppendRow(sheet,
			p.Name,
			p.Ironium,
			p.Boranium,
			p.Germanium,
			p.EnRouteIronium,
			p.EnRouteBoranium,
			p.EnRouteGermanium,
			"", // Total I - will be formula
			"", // Total B - will be formula
			"", // Total G - will be formula
			p.Population,
		)

		// Set formulas for Total columns (H, I, J) = on-planet + en route
		row := doc.RowCount(sheet) - 1
		doc.SetCellFormula(sheet, row, 7, fmt.Sprintf("[.B%d]+[.E%d]", currentRow, currentRow)) // Total I = B + E
		doc.SetCellFormula(sheet, row, 8, fmt.Sprintf("[.C%d]+[.F%d]", currentRow, currentRow)) // Total B = C + F
		doc.SetCellFormula(sheet, row, 9, fmt.Sprintf("[.D%d]+[.G%d]", currentRow, currentRow)) // Total G = D + G
	}

	// Add empty row and totals row with formulas
	if dataRowCount > 0 {
		doc.AppendRow(sheet, "", "", "", "", "", "", "", "", "", "", "")

		endRow := dataStartRow + dataRowCount - 1
		doc.AppendRow(sheet, "TOTAL", "", "", "", "", "", "", "", "", "", "")

		totalRow := doc.RowCount(sheet) - 1
		// Sum formulas for each column
		doc.SetCellFormula(sheet, totalRow, 1, fmt.Sprintf("SUM([.B%d:.B%d])", dataStartRow, endRow))  // Ironium
		doc.SetCellFormula(sheet, totalRow, 2, fmt.Sprintf("SUM([.C%d:.C%d])", dataStartRow, endRow))  // Boranium
		doc.SetCellFormula(sheet, totalRow, 3, fmt.Sprintf("SUM([.D%d:.D%d])", dataStartRow, endRow))  // Germanium
		doc.SetCellFormula(sheet, totalRow, 4, fmt.Sprintf("SUM([.E%d:.E%d])", dataStartRow, endRow))  // En Route I
		doc.SetCellFormula(sheet, totalRow, 5, fmt.Sprintf("SUM([.F%d:.F%d])", dataStartRow, endRow))  // En Route B
		doc.SetCellFormula(sheet, totalRow, 6, fmt.Sprintf("SUM([.G%d:.G%d])", dataStartRow, endRow))  // En Route G
		doc.SetCellFormula(sheet, totalRow, 7, fmt.Sprintf("SUM([.H%d:.H%d])", dataStartRow, endRow))  // Total I
		doc.SetCellFormula(sheet, totalRow, 8, fmt.Sprintf("SUM([.I%d:.I%d])", dataStartRow, endRow))  // Total B
		doc.SetCellFormula(sheet, totalRow, 9, fmt.Sprintf("SUM([.J%d:.J%d])", dataStartRow, endRow))  // Total G
		doc.SetCellFormula(sheet, totalRow, 10, fmt.Sprintf("SUM([.K%d:.K%d])", dataStartRow, endRow)) // Population
	}

	return nil
}

// generateMyMineralsHistorySheet creates/updates the My Minerals History sheet.
func (r *Reporter) generateMyMineralsHistorySheet(doc *ODSDocument, opts *ReportOptions) error {
	sheet := doc.SheetByName(SheetMyMineralHist)
	if sheet == nil {
		return fmt.Errorf("sheet %q not found", SheetMyMineralHist)
	}

	// Always ensure header row exists
	if doc.RowCount(sheet) == 0 || doc.GetCellString(sheet, 0, 0) != "Turn" {
		doc.SetHeaderRow(sheet, "Turn", "Year", "Ironium", "Boranium", "Germanium", "Delta I", "Delta B", "Delta G")
	}

	// Get current data
	snap := r.CollectPlayerSnapshot(opts.PlayerNumber)

	// Check if this turn already exists (avoid duplicates)
	for row := 1; row < doc.RowCount(sheet); row++ {
		turnVal, ok := doc.GetCellInt(sheet, row, 0)
		if ok && turnVal == int64(r.Turn()) {
			// Already have this turn, update it
			doc.SetCellInt(sheet, row, 0, int64(r.Turn()))
			doc.SetCellInt(sheet, row, 1, int64(r.Year()))
			doc.SetCellInt(sheet, row, 2, snap.Minerals.Ironium)
			doc.SetCellInt(sheet, row, 3, snap.Minerals.Boranium)
			doc.SetCellInt(sheet, row, 4, snap.Minerals.Germanium)
			return nil
		}
	}

	// Calculate delta from previous turn
	var deltaI, deltaB, deltaG int64
	if doc.RowCount(sheet) > 1 {
		lastRow := doc.RowCount(sheet) - 1
		if prevI, ok := doc.GetCellInt(sheet, lastRow, 2); ok {
			deltaI = snap.Minerals.Ironium - prevI
		}
		if prevB, ok := doc.GetCellInt(sheet, lastRow, 3); ok {
			deltaB = snap.Minerals.Boranium - prevB
		}
		if prevG, ok := doc.GetCellInt(sheet, lastRow, 4); ok {
			deltaG = snap.Minerals.Germanium - prevG
		}
	}

	// Append new row
	doc.AppendRow(sheet,
		int64(r.Turn()),
		int64(r.Year()),
		snap.Minerals.Ironium,
		snap.Minerals.Boranium,
		snap.Minerals.Germanium,
		deltaI,
		deltaB,
		deltaG,
	)

	return nil
}

// generateMineralShuffleSheet creates the Mineral Shuffle sheet.
func (r *Reporter) generateMineralShuffleSheet(doc *ODSDocument, opts *ReportOptions) error {
	sheet := doc.SheetByName(SheetMineralShuffle)
	if sheet == nil {
		return fmt.Errorf("sheet %q not found", SheetMineralShuffle)
	}

	// Clear and set headers
	doc.ClearSheet(sheet, 0)
	doc.SetHeaderRow(sheet, "From Planet", "To Planet", "Mineral", "Amount", "Distance (ly)")

	// Get recommendations
	recs := r.AnalyzeMineralShuffling(opts.PlayerNumber, opts.MineralThreshold)

	for _, rec := range recs {
		doc.AppendRow(sheet,
			rec.From.Name,
			rec.To.Name,
			rec.Mineral,
			rec.Amount,
			rec.Distance,
		)
	}

	if len(recs) == 0 {
		doc.AppendRow(sheet, "No shuffle recommendations", "", "", "", "")
	}

	return nil
}

// generateOpponentPopulationSheet creates the Opponent Population sheet.
// Shows population on visible planets AND colonists in transit on fleets.
// Only includes data that is actually visible to the player (respects fog of war).
func (r *Reporter) generateOpponentPopulationSheet(doc *ODSDocument, opts *ReportOptions) error {
	sheet := doc.SheetByName(SheetOpponentPop)
	if sheet == nil {
		return fmt.Errorf("sheet %q not found", SheetOpponentPop)
	}

	// Clear and set headers
	doc.ClearSheet(sheet, 0)
	doc.SetHeaderRow(sheet, "Player", "Location", "Type", "Population", "X", "Y", "Last Seen", "Stale")

	currentTurn := r.store.Turn

	// Collect all visible opponent population (planets + fleets with colonists)
	type popInfo struct {
		PlayerName string
		Location   string
		LocType    string // "Planet" or "Fleet"
		Population int64
		X, Y       int
		LastSeen   int
		IsStale    bool
	}
	var popSources []popInfo

	// Add planets with visible population
	for _, p := range r.store.VisibleOpponentPlanets(opts.PlayerNumber) {
		// Only include if we can see population
		if !r.store.CanSeePopulation(p, opts.PlayerNumber) {
			continue
		}
		if p.Population == 0 {
			continue
		}

		playerName := fmt.Sprintf("Player %d", p.Owner)
		if player, ok := r.store.Player(p.Owner); ok {
			playerName = player.NamePlural
		}

		lastSeen := int(p.Meta().Turn)
		popSources = append(popSources, popInfo{
			PlayerName: playerName,
			Location:   p.Name,
			LocType:    "Planet",
			Population: p.Population,
			X:          p.X,
			Y:          p.Y,
			LastSeen:   lastSeen,
			IsStale:    p.Meta().Turn < currentTurn,
		})
	}

	// Add fleets with colonists in cargo
	for _, player := range r.store.AllPlayers() {
		if player.PlayerNumber == opts.PlayerNumber {
			continue
		}

		for _, fleet := range r.store.FleetsByOwner(player.PlayerNumber) {
			cargo := fleet.GetCargo()
			if cargo.Population == 0 {
				continue
			}

			lastSeen := int(fleet.Meta().Turn)
			popSources = append(popSources, popInfo{
				PlayerName: player.NamePlural,
				Location:   fleet.Name(),
				LocType:    "Fleet",
				Population: cargo.Population,
				X:          fleet.X,
				Y:          fleet.Y,
				LastSeen:   lastSeen,
				IsStale:    fleet.Meta().Turn < currentTurn,
			})
		}
	}

	// Sort by population descending
	sort.Slice(popSources, func(i, j int) bool {
		return popSources[i].Population > popSources[j].Population
	})

	for _, p := range popSources {
		staleNote := ""
		if p.IsStale {
			staleNote = fmt.Sprintf("Stale (%d turns old)", r.Turn()-p.LastSeen)
		}
		doc.AppendRow(sheet,
			p.PlayerName,
			p.Location,
			p.LocType,
			p.Population,
			int64(p.X),
			int64(p.Y),
			int64(p.LastSeen),
			staleNote,
		)
	}

	return nil
}

// generateOpponentPopHistorySheet creates/updates the Opponent Pop History sheet.
// Tracks total population (planets + colonists in transit) over time.
func (r *Reporter) generateOpponentPopHistorySheet(doc *ODSDocument, opts *ReportOptions) error {
	sheet := doc.SheetByName(SheetOpponentPopHist)
	if sheet == nil {
		return fmt.Errorf("sheet %q not found", SheetOpponentPopHist)
	}

	// Always ensure header row exists
	if doc.RowCount(sheet) == 0 || doc.GetCellString(sheet, 0, 0) != "Turn" {
		doc.SetHeaderRow(sheet, "Turn", "Year", "Player", "Planet Pop", "Transit Pop", "Total Pop", "Delta Pop")
	}

	// Collect data for all opponents we have population visibility of
	for _, player := range r.store.AllPlayers() {
		if player.PlayerNumber == opts.PlayerNumber {
			continue
		}
		// Skip opponents we have no population visibility of
		if !r.HasPopulationVisibilityOf(player.PlayerNumber, opts.PlayerNumber) {
			continue
		}

		snap := r.CollectVisibleOpponentSnapshot(player.PlayerNumber, opts.PlayerNumber)

		// Calculate colonists in transit
		var transitPop int64
		for _, fleet := range r.store.FleetsByOwner(player.PlayerNumber) {
			transitPop += fleet.GetCargo().Population
		}
		totalPop := snap.TotalPopulation + transitPop

		// Check if this turn+player already exists
		exists := false
		for row := 1; row < doc.RowCount(sheet); row++ {
			turnVal, _ := doc.GetCellInt(sheet, row, 0)
			playerName := doc.GetCellString(sheet, row, 2)
			if turnVal == int64(r.Turn()) && playerName == player.NamePlural {
				exists = true
				break
			}
		}

		if exists {
			continue
		}

		// Calculate delta from previous turn for this player (compare total pop column index 5)
		var deltaPop int64
		for row := doc.RowCount(sheet) - 1; row >= 1; row-- {
			playerName := doc.GetCellString(sheet, row, 2)
			if playerName == player.NamePlural {
				if prevPop, ok := doc.GetCellInt(sheet, row, 5); ok {
					deltaPop = totalPop - prevPop
				}
				break
			}
		}

		doc.AppendRow(sheet,
			int64(r.Turn()),
			int64(r.Year()),
			player.NamePlural,
			snap.TotalPopulation, // Planet pop
			transitPop,           // Transit pop
			totalPop,             // Total pop
			deltaPop,
		)
	}

	return nil
}

// generateOpponentShipsSheet creates/updates the Opponent Ships sheet.
// Only shows ships that are actually visible to the player.
func (r *Reporter) generateOpponentShipsSheet(doc *ODSDocument, opts *ReportOptions) error {
	sheet := doc.SheetByName(SheetOpponentShips)
	if sheet == nil {
		return fmt.Errorf("sheet %q not found", SheetOpponentShips)
	}

	// Always ensure header row exists
	if doc.RowCount(sheet) == 0 || doc.GetCellString(sheet, 0, 0) != "Turn" {
		doc.SetHeaderRow(sheet, "Turn", "Year", "Player", "Unarmed", "Escort", "Capital", "Total", "Delta")
	}

	// Collect data for all opponents we have design visibility of
	for _, player := range r.store.AllPlayers() {
		if player.PlayerNumber == opts.PlayerNumber {
			continue
		}
		// Skip opponents we have no design visibility of (need designs to categorize ships)
		if !r.HasDesignVisibilityOf(player.PlayerNumber) {
			continue
		}

		snap := r.CollectVisibleOpponentSnapshot(player.PlayerNumber, opts.PlayerNumber)

		// Check if this turn+player already exists
		exists := false
		for row := 1; row < doc.RowCount(sheet); row++ {
			turnVal, _ := doc.GetCellInt(sheet, row, 0)
			playerName := doc.GetCellString(sheet, row, 2)
			if turnVal == int64(r.Turn()) && playerName == player.NamePlural {
				exists = true
				break
			}
		}

		if exists {
			continue
		}

		// Calculate delta from previous turn for this player
		var deltaShips int64
		for row := doc.RowCount(sheet) - 1; row >= 1; row-- {
			playerName := doc.GetCellString(sheet, row, 2)
			if playerName == player.NamePlural {
				if prevTotal, ok := doc.GetCellInt(sheet, row, 6); ok {
					deltaShips = int64(snap.Ships.Total) - prevTotal
				}
				break
			}
		}

		doc.AppendRow(sheet,
			int64(r.Turn()),
			int64(r.Year()),
			player.NamePlural,
			int64(snap.Ships.Unarmed),
			int64(snap.Ships.Escort),
			int64(snap.Ships.Capital),
			int64(snap.Ships.Total),
			deltaShips,
		)
	}

	return nil
}

// generateOpponentFleetsSheet creates the Opponent Fleets sheet.
// Shows all visible opponent fleets with position, cargo, and design info.
func (r *Reporter) generateOpponentFleetsSheet(doc *ODSDocument, opts *ReportOptions) error {
	sheet := doc.SheetByName(SheetOpponentFleets)
	if sheet == nil {
		return fmt.Errorf("sheet %q not found", SheetOpponentFleets)
	}

	// Clear and set headers
	doc.ClearSheet(sheet, 0)
	doc.SetHeaderRow(sheet, "Player", "Fleet", "X", "Y", "Warp", "Ships", "Design", "Ironium", "Boranium", "Germanium", "Colonists", "Last Seen", "Stale")

	currentTurn := r.Turn()

	for _, player := range r.store.AllPlayers() {
		if player.PlayerNumber == opts.PlayerNumber {
			continue
		}

		for _, fleet := range r.store.FleetsByOwner(player.PlayerNumber) {
			// Skip empty fleets (ghost entries)
			shipCount := fleet.TotalShips()
			if shipCount == 0 {
				continue
			}

			// Get design info if available
			designInfo := "Unknown"
			designs := fleet.GetDesigns(r.store)
			if len(designs) > 0 {
				var designParts []string
				for _, d := range designs {
					if d.Design != nil && d.Count > 0 {
						designParts = append(designParts, fmt.Sprintf("%s x%d", d.Design.Name, d.Count))
					} else if d.Count > 0 {
						designParts = append(designParts, fmt.Sprintf("? x%d", d.Count))
					}
				}
				if len(designParts) > 0 {
					designInfo = ""
					for i, p := range designParts {
						if i > 0 {
							designInfo += ", "
						}
						designInfo += p
					}
				}
			}

			// Get cargo
			cargo := fleet.GetCargo()

			// Check staleness
			lastSeen := int(fleet.Meta().Turn)
			staleNote := ""
			if lastSeen < currentTurn {
				staleNote = fmt.Sprintf("Stale (%d turns old)", currentTurn-lastSeen)
			}

			// Get warp speed (0 if stationary)
			warp := fleet.Warp

			doc.AppendRow(sheet,
				player.NamePlural,
				fleet.Name(),
				int64(fleet.X),
				int64(fleet.Y),
				int64(warp),
				int64(shipCount),
				designInfo,
				cargo.Ironium,
				cargo.Boranium,
				cargo.Germanium,
				cargo.Population,
				int64(lastSeen),
				staleNote,
			)
		}
	}

	return nil
}

// generateNewDesignsSheet creates the New Designs sheet.
// Shows all known enemy designs with the turn they were first observed.
func (r *Reporter) generateNewDesignsSheet(doc *ODSDocument, opts *ReportOptions) error {
	sheet := doc.SheetByName(SheetNewDesigns)
	if sheet == nil {
		return fmt.Errorf("sheet %q not found", SheetNewDesigns)
	}

	// Clear and set headers
	doc.ClearSheet(sheet, 0)
	doc.SetHeaderRow(sheet, "Player", "Design", "Hull", "Combat Power", "Type", "First Seen")

	currentTurn := r.Turn()

	for _, player := range r.store.AllPlayers() {
		if player.PlayerNumber == opts.PlayerNumber {
			continue
		}
		// Skip opponents we have no design visibility of
		if !r.HasDesignVisibilityOf(player.PlayerNumber) {
			continue
		}

		designs := r.store.ShipDesignsByOwner(player.PlayerNumber)
		for _, design := range designs {
			designType := "Ship"
			if design.IsStarbase {
				designType = "Starbase"
			}

			hullName := ""
			if hull := data.GetHull(design.HullId); hull != nil {
				hullName = hull.Name
			}

			// Get turn when design was first seen
			firstSeen := int(design.Meta().Turn)
			firstSeenNote := fmt.Sprintf("Turn %d", firstSeen)
			if firstSeen == currentTurn {
				firstSeenNote = "NEW"
			}

			doc.AppendRow(sheet,
				player.NamePlural,
				design.Name,
				hullName,
				int64(design.GetCombatPower()),
				designType,
				firstSeenNote,
			)
		}
	}

	return nil
}

// generateScoreEstimatesSheet creates the Score Estimates sheet.
// For the player, shows full score breakdown. For opponents, only shows
// estimates based on visible data (will underestimate true scores).
func (r *Reporter) generateScoreEstimatesSheet(doc *ODSDocument, opts *ReportOptions) error {
	sheet := doc.SheetByName(SheetScoreEstimates)
	if sheet == nil {
		return fmt.Errorf("sheet %q not found", SheetScoreEstimates)
	}

	// Clear and set headers
	doc.ClearSheet(sheet, 0)
	doc.SetHeaderRow(sheet, "Player", "Est Score", "Pop Score", "Resource Score", "Starbase Score", "Tech Score", "Ship Score", "Note")

	for _, player := range r.store.AllPlayers() {
		if player.PlayerNumber == opts.PlayerNumber {
			// Full score for own player
			sc := r.store.ComputeScoreFromActualData(player.PlayerNumber)
			doc.AppendRow(sheet,
				player.NamePlural,
				int64(sc.Score),
				int64(sc.PlanetPopScore),
				int64(sc.ResourceScore),
				int64(sc.StarbaseScore),
				int64(sc.TechScore),
				int64(sc.ShipScore),
				"Full data",
			)
		} else if r.HasVisibilityOf(player.PlayerNumber) {
			// For opponents we have visibility of, use visibility-aware data
			snap := r.CollectVisibleOpponentSnapshot(player.PlayerNumber, opts.PlayerNumber)
			// Simple estimate based on visible data
			popScore := snap.PlanetCount * 2 // Rough estimate
			shipScore := snap.Ships.Total / 10
			doc.AppendRow(sheet,
				player.NamePlural,
				int64(popScore+shipScore), // Very rough estimate
				int64(popScore),
				int64(0), // Can't see resources
				int64(0), // Can't see starbases without DetFull
				int64(0), // Can't see tech
				int64(shipScore),
				"Visible only",
			)
		}
	}

	return nil
}
