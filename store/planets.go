package store

import (
	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
)

// Planetary scanner ID constants
const (
	ScannerNone    = 0  // No scanner installed
	ScannerInvalid = 31 // Special value meaning "no scanner" in file format
)

// PlanetEntity represents a planet with full context.
type PlanetEntity struct {
	meta EntityMeta

	// Core identification
	PlanetNumber int
	Owner        int // -1 = unowned
	IsHomeworld  bool

	// Name (from PlanetsBlock)
	Name string
	X, Y int

	// Detection level (bits 0-6 of flags word)
	// Determines what information is available about this planet.
	// Use blocks.DetNotVisible, DetPenScan, DetSpecial, DetNormalScan, DetFull, DetMaximum constants.
	DetectionLevel int

	// Status flags
	Include          bool // fInclude: planet included in scans/reports
	HasStarbase      bool // fStarbase: planet has a starbase
	HasArtifact      bool // fIsArtifact: planet has an ancient artifact (from header)
	IsTerraformed    bool // fIncEVO: original environment values included
	HasInstallations bool // fIncImp: installations data included
	FirstYear        bool // fFirstYear: first year this planet is visible to player

	// Environment data
	IroniumConc   int
	BoraniumConc  int
	GermaniumConc int
	Gravity       int
	Temperature   int
	Radiation     int

	// Original environment (if terraformed)
	OrigGravity     int
	OrigTemperature int
	OrigRadiation   int

	// Surface minerals
	Ironium   int64
	Boranium  int64
	Germanium int64

	// Installations (from 8-byte installations block if HasInstallations)
	Mines        int  // 12-bit mine count (0-4095)
	Factories    int  // 12-bit factory count (0-4095)
	Defenses     int  // 12-bit defense count (0-4095)
	DeltaPop     int  // 8-bit population change indicator
	ScannerID    int  // 5-bit planetary scanner ID (0=none, 31=no scanner)
	InstArtifact bool // fArtifact from installations dword (bit 22)
	NoResearch   bool // fNoResearch: don't contribute to research (bit 23)
	Population   int64

	// Starbase design slot (if HasStarbase)
	StarbaseDesign int

	// Route (if has route)
	RouteTarget int

	// Raw block (preserved for re-encoding)
	planetBlock *blocks.PartialPlanetBlock
}

// Meta returns the entity metadata.
func (p *PlanetEntity) Meta() *EntityMeta {
	return &p.meta
}

// RawBlocks returns the original blocks.
func (p *PlanetEntity) RawBlocks() []blocks.Block {
	if p.planetBlock != nil {
		return []blocks.Block{*p.planetBlock}
	}
	return nil
}

// SetDirty marks the entity as modified.
func (p *PlanetEntity) SetDirty() {
	p.meta.Dirty = true
}

// IsOwned returns true if the planet has an owner.
func (p *PlanetEntity) IsOwned() bool {
	return p.Owner >= 0
}

// GetMinerals returns the current surface minerals as a Cargo struct.
func (p *PlanetEntity) GetMinerals() Cargo {
	return Cargo{
		Ironium:   p.Ironium,
		Boranium:  p.Boranium,
		Germanium: p.Germanium,
	}
}

// SetMinerals sets surface minerals (marks dirty).
func (p *PlanetEntity) SetMinerals(c Cargo) {
	p.Ironium = c.Ironium
	p.Boranium = c.Boranium
	p.Germanium = c.Germanium
	p.SetDirty()
}

// HasScanner returns true if the planet has a planetary scanner installed.
// ScannerID of 0 or 31 means no scanner.
func (p *PlanetEntity) HasScanner() bool {
	return p.ScannerID > ScannerNone && p.ScannerID < ScannerInvalid
}

// CanSeeEnvironment returns true if environment data (gravity, temp, radiation,
// mineral concentrations) is available for this planet.
// This is true when DetectionLevel >= DetSpecial (2).
func (p *PlanetEntity) CanSeeEnvironment() bool {
	return p.DetectionLevel >= blocks.DetSpecial
}

// DetectionLevelName returns a human-readable name for the detection level.
func (p *PlanetEntity) DetectionLevelName() string {
	switch p.DetectionLevel {
	case blocks.DetNotVisible:
		return "NotVisible"
	case blocks.DetPenScan:
		return "PenScan"
	case blocks.DetSpecial:
		return "Special"
	case blocks.DetNormalScan:
		return "NormalScan"
	case blocks.DetMaximum:
		return "Maximum"
	default:
		if p.DetectionLevel >= blocks.DetFull {
			return "Full"
		}
		return "Unknown"
	}
}

// GetScannerRanges returns the best scanner ranges for this planet.
// This handles:
// - Planetary scanner (based on owner's tech level, if ScannerID is valid)
// - Starbase scanner (if HasStarbase)
// - AR PRT intrinsic scanner (sqrt(population/10), or √2 multiplier with NAS)
// - NAS LRT effects: 2× normal scanner range, no penetrating scanners
// Returns (normal, penetrating) ranges in light-years.
func (p *PlanetEntity) GetScannerRanges(gs *GameStore) (int, int) {
	if p.Owner < 0 {
		return 0, 0
	}

	bestNormal := 0
	bestPen := 0

	player, ok := gs.Player(p.Owner)
	if !ok {
		return 0, 0
	}

	// Check for NAS LRT
	nasLRT := data.GetLRTByCode("NAS")
	hasNAS := nasLRT != nil && player.HasLRT(nasLRT.Bitmask)

	// 1. Planetary scanner (if planet has scanner building)
	if p.HasScanner() {
		scanner, _ := data.GetBestPlanetaryScanner(player.Tech)
		if scanner != nil {
			scanNormal := scanner.NormalRange
			scanPen := scanner.PenetratingRange

			// Apply NAS effects: 2× normal range, no penetrating
			if hasNAS {
				scanNormal *= nasLRT.NormalScannerMultiplier
				scanPen = 0
			}

			if scanNormal > bestNormal {
				bestNormal = scanNormal
			}
			if scanPen > bestPen {
				bestPen = scanPen
			}
		}
	}

	// 2. Starbase scanner (if planet has starbase)
	if p.HasStarbase {
		if starbase, ok := gs.StarbaseDesign(p.Owner, p.StarbaseDesign); ok {
			sbNormal, sbPen := starbase.GetScannerRanges()

			// Apply NAS effects: 2× normal range, no penetrating
			if hasNAS {
				sbNormal *= nasLRT.NormalScannerMultiplier
				sbPen = 0
			}

			if sbNormal > bestNormal {
				bestNormal = sbNormal
			}
			if sbPen > bestPen {
				bestPen = sbPen
			}
		}

		// 3. AR PRT intrinsic scanner
		if prt := data.GetPRT(player.PRT); prt != nil && prt.HasIntrinsicScanner {
			var arRange int
			if hasNAS {
				// AR + NAS: use √2 multiplier formula
				arRange = data.ARNASScannerRange(p.Population)
			} else {
				// AR without NAS: normal formula sqrt(population/10)
				arRange = prt.IntrinsicScannerRange(p.Population)
			}
			if arRange > bestNormal {
				bestNormal = arRange
			}
			// AR intrinsic scanner is normal-only, no penetrating
		}
	}

	return bestNormal, bestPen
}

// qualityFromPlanetBlock determines data quality from planet flags.
func qualityFromPlanetBlock(pb *blocks.PartialPlanetBlock) DataQuality {
	// Full planet data includes installations and population
	if pb.HasInstallations {
		return QualityFull
	}
	// Has environment info (detection level >= DetSpecial) but not full data
	if pb.DetectionLevel >= blocks.DetSpecial {
		return QualityPartial
	}
	// Minimal - just position and owner
	return QualityMinimal
}

// newPlanetEntityFromBlock creates a PlanetEntity from a PartialPlanetBlock.
func newPlanetEntityFromBlock(pb *blocks.PartialPlanetBlock, source *FileSource) *PlanetEntity {
	entity := &PlanetEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   EntityTypePlanet,
				Owner:  pb.Owner,
				Number: pb.PlanetNumber,
			},
			BestSource: source,
			Quality:    qualityFromPlanetBlock(pb),
			Turn:       source.Turn,
		},
		PlanetNumber:     pb.PlanetNumber,
		Owner:            pb.Owner,
		IsHomeworld:      pb.IsHomeworld,
		DetectionLevel:   pb.DetectionLevel,
		Include:          pb.Include,
		HasStarbase:      pb.HasStarbase,
		HasArtifact:      pb.HasArtifact,
		IsTerraformed:    pb.IsTerraformed,
		HasInstallations: pb.HasInstallations,
		FirstYear:        pb.FirstYear,
		IroniumConc:      pb.IroniumConc,
		BoraniumConc:     pb.BoraniumConc,
		GermaniumConc:    pb.GermaniumConc,
		Gravity:          pb.Gravity,
		Temperature:      pb.Temperature,
		Radiation:        pb.Radiation,
		OrigGravity:      pb.OrigGravity,
		OrigTemperature:  pb.OrigTemperature,
		OrigRadiation:    pb.OrigRadiation,
		Ironium:          pb.Ironium,
		Boranium:         pb.Boranium,
		Germanium:        pb.Germanium,
		Mines:            pb.Mines,
		Factories:        pb.Factories,
		Defenses:         pb.Defenses,
		DeltaPop:         pb.DeltaPop,
		ScannerID:        pb.ScannerID,
		InstArtifact:     pb.InstArtifact,
		NoResearch:       pb.NoResearch,
		// Population in Stars! files is stored in 100s of colonists
		Population:     pb.Population * 100,
		StarbaseDesign: pb.StarbaseDesign,
		RouteTarget:    pb.RouteTarget,
		planetBlock:    pb,
	}
	entity.meta.AddSource(source)
	return entity
}

// SetPopulation sets the planet's population (in actual colonists, not file units).
func (p *PlanetEntity) SetPopulation(pop int64) {
	p.Population = pop
	p.SetDirty()
}

// SetInstallations sets mines, factories, and defenses (12-bit values, 0-4095).
// Sets HasInstallations flag based on whether any installation value is non-zero.
func (p *PlanetEntity) SetInstallations(mines, factories, defenses int) {
	p.Mines = mines
	p.Factories = factories
	p.Defenses = defenses
	// HasInstallations if any value is set, including scanner
	p.HasInstallations = mines > 0 || factories > 0 || defenses > 0 || p.HasScanner()
	p.SetDirty()
}

// SetScannerID sets the planet's scanner type index (0=none, 1-30=valid, 31=no scanner).
func (p *PlanetEntity) SetScannerID(scannerID int) {
	p.ScannerID = scannerID
	// If setting a valid scanner, ensure HasInstallations is true
	if scannerID > ScannerNone && scannerID < ScannerInvalid {
		p.HasInstallations = true
	}
	p.SetDirty()
}

// SetMineralConcentrations sets iron, boranium, and germanium concentrations (0-100).
func (p *PlanetEntity) SetMineralConcentrations(iron, boran, germ int) {
	p.IroniumConc = iron
	p.BoraniumConc = boran
	p.GermaniumConc = germ
	p.SetDirty()
}

// SetHabitability sets gravity, temperature, and radiation (0-100 internal scale).
func (p *PlanetEntity) SetHabitability(gravity, temperature, radiation int) {
	p.Gravity = gravity
	p.Temperature = temperature
	p.Radiation = radiation
	p.SetDirty()
}

// MaxPopulation returns the maximum population this planet can support for the given race.
// This delegates to GameStore.MaxPopulation.
func (p *PlanetEntity) MaxPopulation(gs *GameStore, player *PlayerEntity) int {
	return gs.MaxPopulation(p, player)
}

// MaxFactories returns the maximum factories this planet can support for the given race.
// This delegates to GameStore.MaxFactories.
func (p *PlanetEntity) MaxFactories(gs *GameStore, player *PlayerEntity) int {
	return gs.MaxFactories(p, player)
}

// MaxMines returns the maximum mines this planet can support for the given race.
// Formula: max(10, (MaxPopulation × MinesOperate) / 10000)
// Where MaxPopulation is in actual colonists and MinesOperate is per 10k colonists.
func (p *PlanetEntity) MaxMines(gs *GameStore, player *PlayerEntity) int {
	if player.PRT == blocks.PRTAlternateReality {
		return 0 // AR races can't have mines
	}
	maxPop := p.MaxPopulation(gs, player)
	minesOperate := player.Production.MinesOperate
	maxMines := maxPop * minesOperate / 10000
	if maxMines < 10 {
		maxMines = 10
	}
	return maxMines
}

// MaxDefenses returns the absolute maximum defenses based on planet habitability.
// This delegates to GameStore.MaxDefenses.
// Formula: clamp(habitability% * 4, 10, 100), AR races return 0.
func (p *PlanetEntity) MaxDefenses(gs *GameStore, player *PlayerEntity) int {
	return gs.MaxDefenses(p, player)
}

// MaxOperableDefenses returns the population-limited defenses that can operate.
// This delegates to GameStore.MaxOperableDefenses.
// This is the actual limit on functional defenses, considering both habitability and population.
func (p *PlanetEntity) MaxOperableDefenses(gs *GameStore, player *PlayerEntity) int {
	return gs.MaxOperableDefenses(p, player)
}

// HabitabilityValue returns the habitability percentage for the given race.
// Positive = habitable (0-100+), negative = uninhabitable (penalty up to -45).
// This delegates to GameStore.PctPlanetDesirability.
func (p *PlanetEntity) HabitabilityValue(gs *GameStore, player *PlayerEntity) int {
	return gs.PctPlanetDesirability(p, player)
}
