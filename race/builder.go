package race


// ValidationError describes a validation failure.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// TransientRace is returned after each builder step with current state and validation.
type TransientRace struct {
	Race     *Race             // Current race configuration (clone, safe to inspect)
	Points   int               // Available advantage points (can be negative)
	IsValid  bool              // True if race is valid (points >= 0 and no errors)
	Errors   []ValidationError // Validation errors
	Warnings []string          // Non-fatal warnings
}

// Builder creates races with a fluent API and real-time validation.
type Builder struct {
	race *Race
}

// New creates a new race builder with default Humanoid values.
func New() *Builder {
	return &Builder{
		race: Default(),
	}
}

// recalculate computes points and validation, returning a TransientRace.
func (b *Builder) recalculate() *TransientRace {
	errs := Validate(b.race)
	points := CalculatePoints(b.race)

	isValid := len(errs) == 0 && points >= 0

	// Add warning if points are negative
	var warnings []string
	if points < 0 {
		warnings = append(warnings, "Race has negative advantage points")
	}

	return &TransientRace{
		Race:     b.race.Clone(),
		Points:   points,
		IsValid:  isValid,
		Errors:   errs,
		Warnings: warnings,
	}
}

// --- Identity ---

// Name sets the singular and plural race names.
func (b *Builder) Name(singular, plural string) *TransientRace {
	b.race.SingularName = singular
	b.race.PluralName = plural
	return b.recalculate()
}

// Password sets the race password.
func (b *Builder) Password(password string) *TransientRace {
	b.race.Password = password
	return b.recalculate()
}

// Icon sets the race icon/logo (0-31).
func (b *Builder) Icon(icon int) *TransientRace {
	b.race.Icon = icon
	return b.recalculate()
}

// --- Traits ---

// PRT sets the Primary Race Trait (0-9).
func (b *Builder) PRT(prt int) *TransientRace {
	b.race.PRT = prt
	return b.recalculate()
}

// AddLRT adds a Lesser Race Trait by index (0-13).
func (b *Builder) AddLRT(lrtIndex int) *TransientRace {
	if lrtIndex >= 0 && lrtIndex < 14 {
		b.race.LRT |= (1 << lrtIndex)
	}
	return b.recalculate()
}

// RemoveLRT removes a Lesser Race Trait by index (0-13).
func (b *Builder) RemoveLRT(lrtIndex int) *TransientRace {
	if lrtIndex >= 0 && lrtIndex < 14 {
		b.race.LRT &^= (1 << lrtIndex)
	}
	return b.recalculate()
}

// SetLRTs sets all Lesser Race Traits from a bitmask.
func (b *Builder) SetLRTs(lrtBitmask uint16) *TransientRace {
	b.race.LRT = lrtBitmask
	return b.recalculate()
}

// --- Habitability ---

// GravityCenter sets the ideal gravity value (0-100).
func (b *Builder) GravityCenter(center int) *TransientRace {
	b.race.GravityCenter = center
	return b.recalculate()
}

// GravityWidth sets the gravity tolerance half-range (0-50).
// The habitable range is center ± width.
func (b *Builder) GravityWidth(width int) *TransientRace {
	b.race.GravityWidth = width
	return b.recalculate()
}

// GravityImmune sets gravity immunity.
// "Immune to Gravity" checkbox.
func (b *Builder) GravityImmune(immune bool) *TransientRace {
	b.race.GravityImmune = immune
	return b.recalculate()
}

// TemperatureCenter sets the ideal temperature value (0-100).
func (b *Builder) TemperatureCenter(center int) *TransientRace {
	b.race.TemperatureCenter = center
	return b.recalculate()
}

// TemperatureWidth sets the temperature tolerance half-range (0-50).
// The habitable range is center ± width.
func (b *Builder) TemperatureWidth(width int) *TransientRace {
	b.race.TemperatureWidth = width
	return b.recalculate()
}

// TemperatureImmune sets temperature immunity.
// "Immune to Temperature" checkbox.
func (b *Builder) TemperatureImmune(immune bool) *TransientRace {
	b.race.TemperatureImmune = immune
	return b.recalculate()
}

// RadiationCenter sets the ideal radiation value (0-100).
func (b *Builder) RadiationCenter(center int) *TransientRace {
	b.race.RadiationCenter = center
	return b.recalculate()
}

// RadiationWidth sets the radiation tolerance half-range (0-50).
// The habitable range is center ± width.
func (b *Builder) RadiationWidth(width int) *TransientRace {
	b.race.RadiationWidth = width
	return b.recalculate()
}

// RadiationImmune sets radiation immunity.
// "Immune to Radiation" checkbox.
func (b *Builder) RadiationImmune(immune bool) *TransientRace {
	b.race.RadiationImmune = immune
	return b.recalculate()
}

// Gravity sets the gravity center and width at once (convenience method).
func (b *Builder) Gravity(center, width int) *TransientRace {
	b.race.GravityCenter = center
	b.race.GravityWidth = width
	return b.recalculate()
}

// Temperature sets the temperature center and width at once (convenience method).
func (b *Builder) Temperature(center, width int) *TransientRace {
	b.race.TemperatureCenter = center
	b.race.TemperatureWidth = width
	return b.recalculate()
}

// Radiation sets the radiation center and width at once (convenience method).
func (b *Builder) Radiation(center, width int) *TransientRace {
	b.race.RadiationCenter = center
	b.race.RadiationWidth = width
	return b.recalculate()
}

// --- Growth ---

// GrowthRate sets the maximum colonist growth rate per year (1-20, representing 1%-20%).
// "Maximum colonist growth rate per year: X%"
func (b *Builder) GrowthRate(rate int) *TransientRace {
	b.race.GrowthRate = rate
	return b.recalculate()
}

// --- Economy ---

// ColonistsPerResource sets how many colonists generate one resource (700-2500).
// "One resource is generated each year for every X colonists."
func (b *Builder) ColonistsPerResource(cpr int) *TransientRace {
	b.race.ColonistsPerResource = cpr
	return b.recalculate()
}

// FactoryOutput sets resources produced per 10 factories (5-15).
// "Every 10 factories produce X resources each year."
func (b *Builder) FactoryOutput(output int) *TransientRace {
	b.race.FactoryOutput = output
	return b.recalculate()
}

// FactoryCost sets resources required to build one factory (5-25).
// "Factories require X resources to build."
func (b *Builder) FactoryCost(cost int) *TransientRace {
	b.race.FactoryCost = cost
	return b.recalculate()
}

// FactoryCount sets max factories per 10,000 colonists (5-25).
// "Every 10,000 colonists may operate up to X factories."
func (b *Builder) FactoryCount(count int) *TransientRace {
	b.race.FactoryCount = count
	return b.recalculate()
}

// FactoriesUseLessGerm sets whether factories cost 1kT less Germanium.
// "Factories cost 1kT less of Germanium to build"
func (b *Builder) FactoriesUseLessGerm(useLess bool) *TransientRace {
	b.race.FactoriesUseLessGerm = useLess
	return b.recalculate()
}

// MineOutput sets kT of each mineral produced per 10 mines (5-25).
// "Every 10 mines produce up to X kT of each mineral every year."
func (b *Builder) MineOutput(output int) *TransientRace {
	b.race.MineOutput = output
	return b.recalculate()
}

// MineCost sets resources required to build one mine (2-15).
// "Mines require X resources to build."
func (b *Builder) MineCost(cost int) *TransientRace {
	b.race.MineCost = cost
	return b.recalculate()
}

// MineCount sets max mines per 10,000 colonists (5-25).
// "Every 10,000 colonists may operate up to X mines."
func (b *Builder) MineCount(count int) *TransientRace {
	b.race.MineCount = count
	return b.recalculate()
}

// Factories sets all factory parameters at once (convenience method).
func (b *Builder) Factories(output, cost, count int, useLessGerm bool) *TransientRace {
	b.race.FactoryOutput = output
	b.race.FactoryCost = cost
	b.race.FactoryCount = count
	b.race.FactoriesUseLessGerm = useLessGerm
	return b.recalculate()
}

// Mines sets all mine parameters at once (convenience method).
func (b *Builder) Mines(output, cost, count int) *TransientRace {
	b.race.MineOutput = output
	b.race.MineCost = cost
	b.race.MineCount = count
	return b.recalculate()
}

// --- Research ---

// ResearchEnergy sets the Energy research cost level.
// Use ResearchCostExtra (0), ResearchCostStandard (1), or ResearchCostLess (2).
func (b *Builder) ResearchEnergy(cost int) *TransientRace {
	b.race.ResearchEnergy = cost
	return b.recalculate()
}

// ResearchWeapons sets the Weapons research cost level.
// Use ResearchCostExtra (0), ResearchCostStandard (1), or ResearchCostLess (2).
func (b *Builder) ResearchWeapons(cost int) *TransientRace {
	b.race.ResearchWeapons = cost
	return b.recalculate()
}

// ResearchPropulsion sets the Propulsion research cost level.
// Use ResearchCostExtra (0), ResearchCostStandard (1), or ResearchCostLess (2).
func (b *Builder) ResearchPropulsion(cost int) *TransientRace {
	b.race.ResearchPropulsion = cost
	return b.recalculate()
}

// ResearchConstruction sets the Construction research cost level.
// Use ResearchCostExtra (0), ResearchCostStandard (1), or ResearchCostLess (2).
func (b *Builder) ResearchConstruction(cost int) *TransientRace {
	b.race.ResearchConstruction = cost
	return b.recalculate()
}

// ResearchElectronics sets the Electronics research cost level.
// Use ResearchCostExtra (0), ResearchCostStandard (1), or ResearchCostLess (2).
func (b *Builder) ResearchElectronics(cost int) *TransientRace {
	b.race.ResearchElectronics = cost
	return b.recalculate()
}

// ResearchBiotech sets the Biotechnology research cost level.
// Use ResearchCostExtra (0), ResearchCostStandard (1), or ResearchCostLess (2).
func (b *Builder) ResearchBiotech(cost int) *TransientRace {
	b.race.ResearchBiotech = cost
	return b.recalculate()
}

// TechsStartHigh sets whether "Costs 75% extra" research fields start at Tech 4.
// "All 'Costs 75% extra' research fields start at Tech 4"
func (b *Builder) TechsStartHigh(startHigh bool) *TransientRace {
	b.race.TechsStartHigh = startHigh
	return b.recalculate()
}

// Research sets research cost levels for all fields at once (convenience method).
// Use ResearchCostExtra (0), ResearchCostStandard (1), or ResearchCostLess (2).
func (b *Builder) Research(energy, weapons, propulsion, construction, electronics, biotech int) *TransientRace {
	b.race.ResearchEnergy = energy
	b.race.ResearchWeapons = weapons
	b.race.ResearchPropulsion = propulsion
	b.race.ResearchConstruction = construction
	b.race.ResearchElectronics = electronics
	b.race.ResearchBiotech = biotech
	return b.recalculate()
}

// --- Leftover Points ---

// LeftoverPointsOn sets where to spend leftover advantage points.
func (b *Builder) LeftoverPointsOn(option LeftoverPointsOption) *TransientRace {
	b.race.LeftoverPointsOn = option
	return b.recalculate()
}

// --- Finalization ---

// Get returns the current TransientRace state without modifying the race.
func (b *Builder) Get() *TransientRace {
	return b.recalculate()
}

// Finish finalizes the race and returns it if valid.
// Returns an error if the race has validation errors or negative points.
func (b *Builder) Finish() (*Race, error) {
	errs := Validate(b.race, true)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	return b.race.Clone(), nil
}
