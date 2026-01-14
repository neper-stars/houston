package reporter

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"

	"github.com/neper-stars/houston/store"
)

// Reporter generates analysis reports from Stars! game data.
type Reporter struct {
	store *store.GameStore

	// Template for new reports
	templateData []byte

	// Existing report data (for history preservation)
	existingReport []byte
}

// New creates a new Reporter with a fresh GameStore.
func New() *Reporter {
	return &Reporter{
		store: store.New(),
	}
}

// NewFromStore creates a Reporter from an existing GameStore.
func NewFromStore(gs *store.GameStore) *Reporter {
	return &Reporter{
		store: gs,
	}
}

// Store returns the underlying GameStore.
func (r *Reporter) Store() *store.GameStore {
	return r.store
}

// SetTemplateBytes sets the ODS template data.
func (r *Reporter) SetTemplateBytes(data []byte) {
	r.templateData = data
}

// SetTemplateFile loads the ODS template from a file.
func (r *Reporter) SetTemplateFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}
	r.templateData = data
	return nil
}

// SetExistingReportBytes sets existing report data (for history preservation).
func (r *Reporter) SetExistingReportBytes(data []byte) {
	r.existingReport = data
}

// SetExistingReportFile loads existing report from a file.
func (r *Reporter) SetExistingReportFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// No existing report is fine
			return nil
		}
		return fmt.Errorf("failed to read existing report: %w", err)
	}
	r.existingReport = data
	return nil
}

// LoadFile loads game data from a Stars! file.
func (r *Reporter) LoadFile(filename string) error {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return r.LoadBytes(filename, fileBytes)
}

// LoadReader loads game data from an io.Reader.
func (r *Reporter) LoadReader(name string, reader io.Reader) error {
	fileBytes, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}
	return r.LoadBytes(name, fileBytes)
}

// LoadBytes loads game data from file bytes.
func (r *Reporter) LoadBytes(name string, fileBytes []byte) error {
	return r.store.AddFile(name, fileBytes)
}

// LoadFileWithXY loads a game file and automatically loads the companion XY file.
func (r *Reporter) LoadFileWithXY(filename string) error {
	return r.store.AddFileWithXY(filename)
}

// GameID returns the game ID.
func (r *Reporter) GameID() uint32 {
	return r.store.GameID
}

// Turn returns the current turn number.
func (r *Reporter) Turn() int {
	return int(r.store.Turn)
}

// Year returns the current year (2400 + turn).
func (r *Reporter) Year() int {
	return 2400 + r.Turn()
}

// DetectedPlayerNumber returns the player number from the first M-file loaded.
// Returns -1 if no M-file was loaded.
func (r *Reporter) DetectedPlayerNumber() int {
	for _, source := range r.store.Sources() {
		if source.Type == store.SourceTypeMFile {
			return source.PlayerIndex
		}
	}
	return -1
}

// GenerateReport creates the report and returns it as bytes.
func (r *Reporter) GenerateReport(opts *ReportOptions) ([]byte, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Load base document (existing report or template)
	var doc *ODSDocument
	var err error

	switch {
	case r.existingReport != nil:
		doc, err = LoadBytes(r.existingReport)
		if err != nil {
			return nil, fmt.Errorf("failed to load existing report: %w", err)
		}
		// Validate game ID matches
		if existingGameID := extractGameIDFromDoc(doc); existingGameID > 0 {
			if existingGameID != r.GameID() {
				_ = doc.Close()
				return nil, fmt.Errorf("game ID mismatch: existing report is for game %d, but loaded data is for game %d", existingGameID, r.GameID())
			}
		}
	case r.templateData != nil:
		doc, err = LoadBytes(r.templateData)
		if err != nil {
			return nil, fmt.Errorf("failed to load template: %w", err)
		}
	default:
		return nil, fmt.Errorf("no template or existing report provided")
	}
	defer func() { _ = doc.Close() }()

	// Generate all sheets
	if err := r.generateSummarySheet(doc, opts); err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	if err := r.generateMyMineralsSheet(doc, opts); err != nil {
		return nil, fmt.Errorf("failed to generate my minerals: %w", err)
	}

	if err := r.generateMyMineralsHistorySheet(doc, opts); err != nil {
		return nil, fmt.Errorf("failed to generate my minerals history: %w", err)
	}

	if err := r.generateMineralShuffleSheet(doc, opts); err != nil {
		return nil, fmt.Errorf("failed to generate mineral shuffle: %w", err)
	}

	if err := r.generateOpponentPopulationSheet(doc, opts); err != nil {
		return nil, fmt.Errorf("failed to generate opponent population: %w", err)
	}

	if err := r.generateOpponentPopHistorySheet(doc, opts); err != nil {
		return nil, fmt.Errorf("failed to generate opponent pop history: %w", err)
	}

	if err := r.generateOpponentShipsSheet(doc, opts); err != nil {
		return nil, fmt.Errorf("failed to generate opponent ships: %w", err)
	}

	if err := r.generateOpponentFleetsSheet(doc, opts); err != nil {
		return nil, fmt.Errorf("failed to generate opponent fleets: %w", err)
	}

	if err := r.generateNewDesignsSheet(doc, opts); err != nil {
		return nil, fmt.Errorf("failed to generate new designs: %w", err)
	}

	if err := r.generateScoreEstimatesSheet(doc, opts); err != nil {
		return nil, fmt.Errorf("failed to generate score estimates: %w", err)
	}

	return doc.WriteBytes()
}

// GenerateReportToFile creates the report and saves it to a file.
func (r *Reporter) GenerateReportToFile(filename string, opts *ReportOptions) error {
	data, err := r.GenerateReport(opts)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

// extractGameIDFromDoc reads the game ID from an existing ODS report.
// Returns 0 if the game ID cannot be found.
func extractGameIDFromDoc(doc *ODSDocument) uint32 {
	sheet := doc.SheetByName(SheetSummary)
	if sheet == nil {
		return 0
	}
	// Game ID is stored in row 1 (after header), column 1 (value column)
	// Row 0: "Game Info", "Value" (header)
	// Row 1: "Game ID", <value>
	if label := doc.GetCellString(sheet, 1, 0); label != "Game ID" {
		return 0
	}
	if gameID, ok := doc.GetCellInt(sheet, 1, 1); ok && gameID > 0 {
		return uint32(gameID) // #nosec G115 -- gameID is validated > 0
	}
	return 0
}

// CollectPlayerSnapshot gathers all data for a player at the current turn.
func (r *Reporter) CollectPlayerSnapshot(playerNumber int) PlayerSnapshot {
	snap := PlayerSnapshot{
		PlayerNumber: playerNumber,
		Turn:         r.Turn(),
		Year:         r.Year(),
	}

	// Get planets owned by this player
	planets := r.store.PlanetsByOwner(playerNumber)
	snap.PlanetCount = len(planets)

	for _, p := range planets {
		snap.TotalPopulation += p.Population
		snap.Minerals.Ironium += p.Ironium
		snap.Minerals.Boranium += p.Boranium
		snap.Minerals.Germanium += p.Germanium
	}

	// Get fleets owned by this player
	fleets := r.store.FleetsByOwner(playerNumber)
	snap.FleetCount = len(fleets)

	// Count ships by category
	for _, fleet := range fleets {
		designs := fleet.GetDesigns(r.store)
		for _, entry := range designs {
			if entry.Design == nil {
				continue
			}
			power := entry.Design.GetCombatPower()
			switch {
			case power == 0:
				snap.Ships.Unarmed += entry.Count
			case power < 2000:
				snap.Ships.Escort += entry.Count
			default:
				snap.Ships.Capital += entry.Count
			}
			snap.Ships.Total += entry.Count
		}
	}

	// Get score estimate
	sc := r.store.ComputeScoreFromActualData(playerNumber)
	snap.EstimatedScore = sc.Score

	// Get tech levels if available
	if player, ok := r.store.Player(playerNumber); ok {
		snap.TechEnergy = player.Tech.Energy
		snap.TechWeapons = player.Tech.Weapons
		snap.TechPropulsion = player.Tech.Propulsion
		snap.TechConstruction = player.Tech.Construction
		snap.TechElectronics = player.Tech.Electronics
		snap.TechBiotech = player.Tech.Biotech
	}

	return snap
}

// CollectVisibleOpponentSnapshot gathers data for an opponent player,
// respecting fog of war (only including data the viewer can actually see).
func (r *Reporter) CollectVisibleOpponentSnapshot(opponentNumber, viewerNumber int) PlayerSnapshot {
	snap := PlayerSnapshot{
		PlayerNumber: opponentNumber,
		Turn:         r.Turn(),
		Year:         r.Year(),
	}

	// Only count visible planets owned by this opponent
	for _, p := range r.store.VisiblePlanetsByOwner(opponentNumber) {
		snap.PlanetCount++
		// Only add population/minerals if detection level allows
		if r.store.CanSeePopulation(p, viewerNumber) {
			snap.TotalPopulation += p.Population
		}
		if r.store.CanSeeMinerals(p, viewerNumber) {
			snap.Minerals.Ironium += p.Ironium
			snap.Minerals.Boranium += p.Boranium
			snap.Minerals.Germanium += p.Germanium
		}
	}

	// Get visible fleets owned by this player
	// (fleets are only in the M file if visible, so this is already filtered)
	fleets := r.store.FleetsByOwner(opponentNumber)
	snap.FleetCount = len(fleets)

	// Count ships by category
	for _, fleet := range fleets {
		designs := fleet.GetDesigns(r.store)
		for _, entry := range designs {
			if entry.Design == nil {
				continue
			}
			power := entry.Design.GetCombatPower()
			switch {
			case power == 0:
				snap.Ships.Unarmed += entry.Count
			case power < 2000:
				snap.Ships.Escort += entry.Count
			default:
				snap.Ships.Capital += entry.Count
			}
			snap.Ships.Total += entry.Count
		}
	}

	// Score estimate based on visible data only
	// Note: This will underestimate since we can't see all planets
	snap.EstimatedScore = 0 // Can't reliably estimate without full data

	// Tech levels are typically not visible for opponents
	// (would need to infer from seen ship designs)

	return snap
}

// HasVisibilityOf returns true if we have any visibility data for an opponent.
// This checks for visible planets, fleets with ships, or ship designs.
func (r *Reporter) HasVisibilityOf(opponentNumber int) bool {
	return r.HasPlanetVisibilityOf(opponentNumber) ||
		r.HasFleetVisibilityOf(opponentNumber) ||
		r.HasDesignVisibilityOf(opponentNumber)
}

// HasPlanetVisibilityOf returns true if we can see any planets owned by opponent.
func (r *Reporter) HasPlanetVisibilityOf(opponentNumber int) bool {
	return len(r.store.VisiblePlanetsByOwner(opponentNumber)) > 0
}

// HasPopulationVisibilityOf returns true if we can see any population for opponent
// (either on planets or as colonists in transit on fleets).
func (r *Reporter) HasPopulationVisibilityOf(opponentNumber, viewerNumber int) bool {
	// Check planets with visible population
	for _, p := range r.store.VisiblePlanetsByOwner(opponentNumber) {
		if r.store.CanSeePopulation(p, viewerNumber) && p.Population > 0 {
			return true
		}
	}
	// Check fleets with colonists
	for _, fleet := range r.store.FleetsByOwner(opponentNumber) {
		if fleet.GetCargo().Population > 0 {
			return true
		}
	}
	return false
}

// HasFleetVisibilityOf returns true if we can see any fleets with ships owned by opponent.
func (r *Reporter) HasFleetVisibilityOf(opponentNumber int) bool {
	for _, fleet := range r.store.FleetsByOwner(opponentNumber) {
		if fleet.TotalShips() > 0 {
			return true
		}
	}
	return false
}

// HasDesignVisibilityOf returns true if we know any ship designs for opponent.
func (r *Reporter) HasDesignVisibilityOf(opponentNumber int) bool {
	return len(r.store.ShipDesignsByOwner(opponentNumber)) > 0
}

// CollectPlanetMineralData gathers mineral data for all planets owned by a player.
func (r *Reporter) CollectPlanetMineralData(playerNumber int) []PlanetMineralData {
	planets := r.store.PlanetsByOwner(playerNumber)
	fleets := r.store.FleetsByOwner(playerNumber)

	var result []PlanetMineralData

	for _, p := range planets {
		data := PlanetMineralData{
			Number:     p.PlanetNumber,
			Name:       p.Name,
			X:          p.X,
			Y:          p.Y,
			Ironium:    p.Ironium,
			Boranium:   p.Boranium,
			Germanium:  p.Germanium,
			Population: p.Population,
		}

		// Find fleets heading to this planet
		for _, fleet := range fleets {
			if len(fleet.Waypoints) > 0 {
				// Check first waypoint destination
				wp := fleet.Waypoints[0]
				if wp.X == p.X && wp.Y == p.Y {
					cargo := fleet.GetCargo()
					data.EnRouteIronium += cargo.Ironium
					data.EnRouteBoranium += cargo.Boranium
					data.EnRouteGermanium += cargo.Germanium
				}
			}
		}

		result = append(result, data)
	}

	// Sort by planet name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// AnalyzeMineralShuffling returns recommendations for moving minerals between planets.
func (r *Reporter) AnalyzeMineralShuffling(playerNumber int, threshold int64) []ShuffleRecommendation {
	planets := r.store.PlanetsByOwner(playerNumber)

	var needs []MineralNeed
	var sources []MineralSource

	// Analyze each planet
	for _, p := range planets {
		// Check Ironium
		if p.Ironium < threshold {
			needs = append(needs, MineralNeed{
				Planet:  p,
				Mineral: "Ironium",
				Current: p.Ironium,
				Needed:  threshold - p.Ironium,
			})
		} else if p.Ironium > threshold*2 {
			sources = append(sources, MineralSource{
				Planet:  p,
				Mineral: "Ironium",
				Surplus: p.Ironium - threshold,
			})
		}

		// Check Boranium
		if p.Boranium < threshold {
			needs = append(needs, MineralNeed{
				Planet:  p,
				Mineral: "Boranium",
				Current: p.Boranium,
				Needed:  threshold - p.Boranium,
			})
		} else if p.Boranium > threshold*2 {
			sources = append(sources, MineralSource{
				Planet:  p,
				Mineral: "Boranium",
				Surplus: p.Boranium - threshold,
			})
		}

		// Check Germanium
		if p.Germanium < threshold {
			needs = append(needs, MineralNeed{
				Planet:  p,
				Mineral: "Germanium",
				Current: p.Germanium,
				Needed:  threshold - p.Germanium,
			})
		} else if p.Germanium > threshold*2 {
			sources = append(sources, MineralSource{
				Planet:  p,
				Mineral: "Germanium",
				Surplus: p.Germanium - threshold,
			})
		}
	}

	// Match needs with sources by mineral type, preferring shortest distances
	var recommendations []ShuffleRecommendation

	for _, need := range needs {
		bestSource := -1
		bestDistance := math.MaxFloat64

		for i, source := range sources {
			if source.Mineral != need.Mineral || source.Surplus <= 0 {
				continue
			}

			dx := float64(source.Planet.X - need.Planet.X)
			dy := float64(source.Planet.Y - need.Planet.Y)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist < bestDistance {
				bestDistance = dist
				bestSource = i
			}
		}

		if bestSource >= 0 {
			source := &sources[bestSource]
			amount := need.Needed
			if amount > source.Surplus {
				amount = source.Surplus
			}

			recommendations = append(recommendations, ShuffleRecommendation{
				From:     source.Planet,
				To:       need.Planet,
				Mineral:  need.Mineral,
				Amount:   amount,
				Distance: bestDistance,
			})

			source.Surplus -= amount
		}
	}

	// Sort by distance
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Distance < recommendations[j].Distance
	})

	return recommendations
}
