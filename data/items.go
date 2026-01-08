// Package data contains static game data and constants for Stars! file parsing.
package data

// ItemCategory represents the category of a ship component.
type ItemCategory int

const (
	CategoryEmpty      ItemCategory = 0
	CategoryOrbital    ItemCategory = 1  // Stargates, Mass Drivers
	CategoryBeamWeapon ItemCategory = 2  // Lasers, Phasers, etc.
	CategoryTorpedo    ItemCategory = 3  // Torpedoes, Missiles
	CategoryBomb       ItemCategory = 4  // Planet bombs
	CategoryTerraform  ItemCategory = 5  // Terraforming devices
	CategoryPlanetary  ItemCategory = 6  // Planetary installations (scanners, defenses)
	CategoryMiningRobo ItemCategory = 7  // Mining robots
	CategoryMineLayer  ItemCategory = 8  // Mine dispensers
	CategoryMechanical ItemCategory = 9  // Cargo pods, fuel tanks, etc.
	CategoryElectrical ItemCategory = 10 // Cloaks, computers, jammers
	CategoryShield     ItemCategory = 11 // Shields
	CategoryScanner    ItemCategory = 12 // Ship scanners
	CategoryArmor      ItemCategory = 13 // Armor
	CategoryEngine     ItemCategory = 14 // Engines
	CategoryShipHull   ItemCategory = 15 // Ship hulls
	CategoryStarbase   ItemCategory = 16 // Starbase hulls
)

// CategoryNames maps category IDs to display names.
var CategoryNames = map[ItemCategory]string{
	CategoryEmpty:      "Empty",
	CategoryOrbital:    "Orbital",
	CategoryBeamWeapon: "Beam Weapon",
	CategoryTorpedo:    "Torpedo",
	CategoryBomb:       "Bomb",
	CategoryTerraform:  "Terraforming",
	CategoryPlanetary:  "Planetary",
	CategoryMiningRobo: "Mining Robot",
	CategoryMineLayer:  "Mine Layer",
	CategoryMechanical: "Mechanical",
	CategoryElectrical: "Electrical",
	CategoryShield:     "Shield",
	CategoryScanner:    "Scanner",
	CategoryArmor:      "Armor",
	CategoryEngine:     "Engine",
	CategoryShipHull:   "Ship Hull",
	CategoryStarbase:   "Starbase Hull",
}

// ItemInfo contains the category and item ID for a component.
type ItemInfo struct {
	Category ItemCategory
	ItemID   int
}

// Engine item IDs
const (
	EngineSettlersDelight         = 1
	EngineQuickJump5              = 2
	EngineFuelMizer               = 3
	EngineLongHump6               = 4
	EngineDaddyLongLegs7          = 5
	EngineAlphaDrive8             = 6
	EngineTransGalacticDrive      = 7
	EngineInterspace10            = 8
	EngineEnigmaPulsar            = 9
	EngineTransStar10             = 10
	EngineRadiatingHydroRamScoop  = 11
	EngineSubGalacticFuelScoop    = 12
	EngineTransGalacticFuelScoop  = 13
	EngineTransGalacticSuperScoop = 14
	EngineTransGalacticMizerScoop = 15
	EngineGalaxyScoop             = 16
)

// Beam weapon item IDs
const (
	BeamLaser                  = 1
	BeamXRayLaser              = 2
	BeamMiniGun                = 3
	BeamYakimoraLightPhaser    = 4
	BeamBlackjack              = 5
	BeamPhaserBazooka          = 6
	BeamPulsedSapper           = 7
	BeamColloidalPhaser        = 8
	BeamGatlingGun             = 9
	BeamMiniBlaster            = 10
	BeamBludgeon               = 11
	BeamMarkIVBlaster          = 12
	BeamPhasedSapper           = 13
	BeamHeavyBlaster           = 14
	BeamGatlingNeutrinoCannon  = 15
	BeamMyopicDisruptor        = 16
	BeamBlunderbuss            = 17
	BeamDisruptor              = 18
	BeamMultiContainedMunition = 19
	BeamSyncroSapper           = 20
	BeamMegaDisruptor          = 21
	BeamBigMuthaCannon         = 22
	BeamStreamingPulverizer    = 23
	BeamAntiMatterPulverizer   = 24
)

// Torpedo item IDs
const (
	TorpedoAlpha      = 1
	TorpedoBeta       = 2
	TorpedoDelta      = 3
	TorpedoEpsilon    = 4
	TorpedoRho        = 5
	TorpedoUpsilon    = 6
	TorpedoOmega      = 7
	TorpedoAntiMatter = 8
	TorpedoJihad      = 9
	TorpedoJuggernaut = 10
	TorpedoDoomsday   = 11
	TorpedoArmageddon = 12
)

// Shield item IDs
const (
	ShieldMoleskin         = 1
	ShieldCowhide          = 2
	ShieldWolverineDiffuse = 3
	ShieldCrobySharmor     = 4
	ShieldShadow           = 5
	ShieldBearNeutrino     = 6
	ShieldLangstonShell    = 7
	ShieldGorillaDelagator = 8
	ShieldElephantHide     = 9
	ShieldCompletePhase    = 10
)

// Scanner item IDs (ship scanners)
const (
	ScannerBat         = 1
	ScannerRhino       = 2
	ScannerMole        = 3
	ScannerDNA         = 4
	ScannerPossum      = 5
	ScannerPickPocket  = 6
	ScannerChameleon   = 7
	ScannerFerret      = 8
	ScannerDolphin     = 9
	ScannerGazelle     = 10
	ScannerRNA         = 11
	ScannerCheetah     = 12
	ScannerElephant    = 13
	ScannerEagleEye    = 14
	ScannerRobberBaron = 15
	ScannerPeerless    = 16
)

// Scanner represents a ship scanner with its stats.
type Scanner struct {
	ID               int
	Name             string
	Tech             TechRequirements
	Mass             int
	Cost             Cost
	NormalRange      int  // Normal scanning range in light-years
	PenetratingRange int  // Penetrating scanning range in light-years
	StealsCargo      bool // Can detect enemy cargo (Pick Pocket, Robber Baron)
	CloakPercent     int  // Cloaking percentage (for Chameleon Scanner)
}

// Scanners contains all ship scanner definitions.
var Scanners = map[int]*Scanner{
	ScannerBat: {
		ID: ScannerBat, Name: "Bat Scanner",
		Tech: TechRequirements{}, Mass: 2, Cost: Cost{1, 1, 0, 1},
		NormalRange: 0, PenetratingRange: 0,
	},
	ScannerRhino: {
		ID: ScannerRhino, Name: "Rhino Scanner",
		Tech: TechRequirements{Electronics: 1}, Mass: 5, Cost: Cost{3, 3, 0, 2},
		NormalRange: 50, PenetratingRange: 0,
	},
	ScannerMole: {
		ID: ScannerMole, Name: "Mole Scanner",
		Tech: TechRequirements{Electronics: 4}, Mass: 2, Cost: Cost{9, 2, 0, 2},
		NormalRange: 100, PenetratingRange: 0,
	},
	ScannerDNA: {
		ID: ScannerDNA, Name: "DNA Scanner",
		Tech: TechRequirements{Propulsion: 3, Biotech: 6}, Mass: 2, Cost: Cost{5, 1, 1, 1},
		NormalRange: 125, PenetratingRange: 0,
	},
	ScannerPossum: {
		ID: ScannerPossum, Name: "Possum Scanner",
		Tech: TechRequirements{Electronics: 5}, Mass: 3, Cost: Cost{18, 3, 0, 3},
		NormalRange: 150, PenetratingRange: 0,
	},
	ScannerPickPocket: {
		ID: ScannerPickPocket, Name: "Pick Pocket Scanner",
		Tech: TechRequirements{Energy: 4, Electronics: 4, Biotech: 4}, Mass: 15, Cost: Cost{35, 8, 10, 6},
		NormalRange: 80, PenetratingRange: 0, StealsCargo: true,
	},
	ScannerChameleon: {
		ID: ScannerChameleon, Name: "Chameleon Scanner",
		Tech: TechRequirements{Energy: 3, Electronics: 6}, Mass: 6, Cost: Cost{25, 4, 6, 4},
		NormalRange: 160, PenetratingRange: 45, CloakPercent: 40,
	},
	ScannerFerret: {
		ID: ScannerFerret, Name: "Ferret Scanner",
		Tech: TechRequirements{Energy: 3, Electronics: 7, Biotech: 2}, Mass: 2, Cost: Cost{36, 2, 0, 8},
		NormalRange: 185, PenetratingRange: 50,
	},
	ScannerDolphin: {
		ID: ScannerDolphin, Name: "Dolphin Scanner",
		Tech: TechRequirements{Energy: 5, Electronics: 10, Biotech: 4}, Mass: 4, Cost: Cost{40, 5, 5, 10},
		NormalRange: 220, PenetratingRange: 100,
	},
	ScannerGazelle: {
		ID: ScannerGazelle, Name: "Gazelle Scanner",
		Tech: TechRequirements{Energy: 4, Electronics: 8}, Mass: 5, Cost: Cost{24, 4, 0, 5},
		NormalRange: 225, PenetratingRange: 0,
	},
	ScannerRNA: {
		ID: ScannerRNA, Name: "RNA Scanner",
		Tech: TechRequirements{Propulsion: 5, Biotech: 10}, Mass: 2, Cost: Cost{20, 1, 1, 2},
		NormalRange: 230, PenetratingRange: 0,
	},
	ScannerCheetah: {
		ID: ScannerCheetah, Name: "Cheetah Scanner",
		Tech: TechRequirements{Energy: 5, Electronics: 11}, Mass: 4, Cost: Cost{50, 3, 1, 13},
		NormalRange: 275, PenetratingRange: 0,
	},
	ScannerElephant: {
		ID: ScannerElephant, Name: "Elephant Scanner",
		Tech: TechRequirements{Energy: 6, Electronics: 16, Biotech: 7}, Mass: 6, Cost: Cost{70, 8, 5, 14},
		NormalRange: 300, PenetratingRange: 200,
	},
	ScannerEagleEye: {
		ID: ScannerEagleEye, Name: "Eagle Eye Scanner",
		Tech: TechRequirements{Energy: 6, Electronics: 14}, Mass: 3, Cost: Cost{64, 3, 2, 21},
		NormalRange: 335, PenetratingRange: 0,
	},
	ScannerRobberBaron: {
		ID: ScannerRobberBaron, Name: "Robber Baron Scanner",
		Tech: TechRequirements{Energy: 10, Electronics: 15, Biotech: 10}, Mass: 20, Cost: Cost{90, 10, 10, 10},
		NormalRange: 220, PenetratingRange: 120, StealsCargo: true,
	},
	ScannerPeerless: {
		ID: ScannerPeerless, Name: "Peerless Scanner",
		Tech: TechRequirements{Energy: 7, Electronics: 24}, Mass: 4, Cost: Cost{90, 3, 2, 30},
		NormalRange: 500, PenetratingRange: 0,
	},
}

// GetScanner returns the scanner for a given ID, or nil if not found.
func GetScanner(id int) *Scanner {
	return Scanners[id]
}

// ScannerStats holds the range capabilities of a scanner.
// Used for planetary scanners and JoAT intrinsic scanners.
type ScannerStats struct {
	NormalRange      int  // Normal scanning range in light-years
	PenetratingRange int  // Penetrating scanning range in light-years
	StealsCargo      bool // Can detect enemy cargo (Pick Pocket, Robber Baron)
	ElectronicsLevel int  // Electronics tech level required (for planetary scanners)
}

// Planetary scanner item IDs
const (
	PlanetaryScannerViewer50   = 1
	PlanetaryScannerViewer90   = 2
	PlanetaryScannerScoper150  = 3
	PlanetaryScannerScoper220  = 4
	PlanetaryScannerScoper280  = 5
	PlanetaryScannerSnooper320 = 6
	PlanetaryScannerSnooper400 = 7
	PlanetaryScannerSnooper500 = 8
	PlanetaryScannerSnooper620 = 9
)

// PlanetaryScanner represents a planetary scanner with its stats.
type PlanetaryScanner struct {
	ID               int
	Name             string
	Tech             TechRequirements
	Cost             Cost
	NormalRange      int // Normal scanning range in light-years
	PenetratingRange int // Penetrating scanning range in light-years
}

// PlanetaryScanners contains all planetary scanner definitions.
var PlanetaryScanners = map[int]*PlanetaryScanner{
	PlanetaryScannerViewer50: {
		ID: PlanetaryScannerViewer50, Name: "Viewer 50",
		Tech: TechRequirements{}, Cost: Cost{100, 10, 10, 70},
		NormalRange: 50, PenetratingRange: 0,
	},
	PlanetaryScannerViewer90: {
		ID: PlanetaryScannerViewer90, Name: "Viewer 90",
		Tech: TechRequirements{Electronics: 1}, Cost: Cost{100, 10, 10, 70},
		NormalRange: 90, PenetratingRange: 0,
	},
	PlanetaryScannerScoper150: {
		ID: PlanetaryScannerScoper150, Name: "Scoper 150",
		Tech: TechRequirements{Electronics: 3}, Cost: Cost{100, 10, 10, 70},
		NormalRange: 150, PenetratingRange: 0,
	},
	PlanetaryScannerScoper220: {
		ID: PlanetaryScannerScoper220, Name: "Scoper 220",
		Tech: TechRequirements{Electronics: 6}, Cost: Cost{100, 10, 10, 70},
		NormalRange: 220, PenetratingRange: 0,
	},
	PlanetaryScannerScoper280: {
		ID: PlanetaryScannerScoper280, Name: "Scoper 280",
		Tech: TechRequirements{Electronics: 8}, Cost: Cost{100, 10, 10, 70},
		NormalRange: 280, PenetratingRange: 0,
	},
	PlanetaryScannerSnooper320: {
		ID: PlanetaryScannerSnooper320, Name: "Snooper 320X",
		Tech: TechRequirements{Energy: 3, Electronics: 10, Biotech: 3}, Cost: Cost{100, 10, 10, 70},
		NormalRange: 320, PenetratingRange: 160,
	},
	PlanetaryScannerSnooper400: {
		ID: PlanetaryScannerSnooper400, Name: "Snooper 400X",
		Tech: TechRequirements{Energy: 4, Electronics: 13, Biotech: 6}, Cost: Cost{100, 10, 10, 70},
		NormalRange: 400, PenetratingRange: 200,
	},
	PlanetaryScannerSnooper500: {
		ID: PlanetaryScannerSnooper500, Name: "Snooper 500X",
		Tech: TechRequirements{Energy: 5, Electronics: 16, Biotech: 7}, Cost: Cost{100, 10, 10, 70},
		NormalRange: 500, PenetratingRange: 250,
	},
	PlanetaryScannerSnooper620: {
		ID: PlanetaryScannerSnooper620, Name: "Snooper 620X",
		Tech: TechRequirements{Energy: 7, Electronics: 23, Biotech: 9}, Cost: Cost{100, 10, 10, 70},
		NormalRange: 620, PenetratingRange: 310,
	},
}

// GetPlanetaryScanner returns the planetary scanner for a given ID, or nil if not found.
func GetPlanetaryScanner(id int) *PlanetaryScanner {
	return PlanetaryScanners[id]
}

// GetBestPlanetaryScanner returns the best planetary scanner available at the given tech levels.
// Returns the scanner definition and the scanner ID.
func GetBestPlanetaryScanner(tech TechRequirements) (*PlanetaryScanner, int) {
	// Ordered from best to worst
	scanners := []int{
		PlanetaryScannerSnooper620,
		PlanetaryScannerSnooper500,
		PlanetaryScannerSnooper400,
		PlanetaryScannerSnooper320,
		PlanetaryScannerScoper280,
		PlanetaryScannerScoper220,
		PlanetaryScannerScoper150,
		PlanetaryScannerViewer90,
		PlanetaryScannerViewer50,
	}

	for _, id := range scanners {
		scanner := PlanetaryScanners[id]
		if scanner.Tech.CanBuildWith(tech) {
			return scanner, id
		}
	}

	// Fallback to Viewer 50 (always available)
	return PlanetaryScanners[PlanetaryScannerViewer50], PlanetaryScannerViewer50
}

// GetShipScannerStats returns the scanner stats for a ship scanner ID.
func GetShipScannerStats(scannerID int) (ScannerStats, bool) {
	s := Scanners[scannerID]
	if s == nil {
		return ScannerStats{}, false
	}
	return ScannerStats{
		NormalRange:      s.NormalRange,
		PenetratingRange: s.PenetratingRange,
		StealsCargo:      s.StealsCargo,
	}, true
}

// GetPlanetaryScannerStats returns the scanner stats for a planetary scanner ID.
func GetPlanetaryScannerStats(scannerID int) (ScannerStats, bool) {
	s := PlanetaryScanners[scannerID]
	if s == nil {
		return ScannerStats{}, false
	}
	return ScannerStats{
		NormalRange:      s.NormalRange,
		PenetratingRange: s.PenetratingRange,
		ElectronicsLevel: s.Tech.Electronics,
	}, true
}

// Armor item IDs
const (
	ArmorTritanium          = 1
	ArmorCrobmnium          = 2
	ArmorCarbonic           = 3
	ArmorStrobnium          = 4
	ArmorOrganic            = 5
	ArmorKelarium           = 6
	ArmorFieldedKelarium    = 7
	ArmorDepletedNeutronium = 8
	ArmorNeutronium         = 9
	ArmorMegaPolyShell      = 10
	ArmorValanium           = 11
	ArmorSuperlatanium      = 12
)

// Mechanical item IDs
// NOTE: IDs updated to match decompiled SpdOfShip() which uses:
//   - Maneuver Jet = 7 (grhst=0x1000)
//   - Overthruster = 8 (grhst=0x1000)
//
// Previous values were MechManeuveringJet=8, MechOverthruster=9 - need test data to validate.
const (
	MechColonizationModule        = 1
	MechOrbitalConstructionModule = 2
	MechCargoPod                  = 3
	MechSuperCargoPod             = 4
	MechMultiCargoPod             = 5
	MechFuelTank                  = 6
	MechManeuveringJet            = 7 // Was 8 - updated per decompiled SpdOfShip()
	MechOverthruster              = 8 // Was 9 - updated per decompiled SpdOfShip()
	MechSuperFuelTank             = 9 // Was 7 - shifted due to Maneuver Jet/Overthruster fix
	MechJumpGate                  = 10
	MechBeamDeflector             = 11
)

// Mining robot item IDs
// Mining robot items (category 0x80).
// NOTE: Decompiled SpdOfShip() hardcodes a +0.5 speed bonus for MiningRoboUltra (ID 6).
// Earlier documentation incorrectly referred to this as "Sub-light Motor" - that was an error.
// There is no separate "Sub-light Motor" item; the speed bonus is an undocumented feature
// of the Robo-Ultra-Miner.
const (
	MiningRoboMidget = iota + 1
	MiningRoboMini
	MiningRobo
	MiningRoboMaxi
	MiningRoboSuper
	MiningRoboUltra // Has hardcoded +0.5 speed bonus in SpdOfShip()
	MiningAlien
	MiningOrbitalAdj
)

// ItemNameToInfo maps item names to their category and ID.
// This allows looking up items by their display name.
var ItemNameToInfo = map[string]ItemInfo{
	// Category 1: Orbital (Stargates, Mass Drivers)
	"Stargate 100/250": {CategoryOrbital, 1},
	"Stargate any/300": {CategoryOrbital, 2},
	"Stargate 150/600": {CategoryOrbital, 3},
	"Stargate 300/500": {CategoryOrbital, 4},
	"Stargate 100/any": {CategoryOrbital, 5},
	"Stargate any/800": {CategoryOrbital, 6},
	"Stargate any/any": {CategoryOrbital, 7},
	"Mass Driver 5":    {CategoryOrbital, 8},
	"Mass Driver 6":    {CategoryOrbital, 9},
	"Mass Driver 7":    {CategoryOrbital, 10},
	"Super Driver 8":   {CategoryOrbital, 11},
	"Super Driver 9":   {CategoryOrbital, 12},
	"Ultra Driver 10":  {CategoryOrbital, 13},
	"Ultra Driver 11":  {CategoryOrbital, 14},
	"Ultra Driver 12":  {CategoryOrbital, 15},
	"Ultra Driver 13":  {CategoryOrbital, 16},

	// Category 2: Beam Weapons
	"Laser":                    {CategoryBeamWeapon, BeamLaser},
	"X-Ray Laser":              {CategoryBeamWeapon, BeamXRayLaser},
	"Mini Gun":                 {CategoryBeamWeapon, BeamMiniGun},
	"Yakimora Light Phaser":    {CategoryBeamWeapon, BeamYakimoraLightPhaser},
	"Blackjack":                {CategoryBeamWeapon, BeamBlackjack},
	"Phaser Bazooka":           {CategoryBeamWeapon, BeamPhaserBazooka},
	"Pulsed Sapper":            {CategoryBeamWeapon, BeamPulsedSapper},
	"Colloidal Phaser":         {CategoryBeamWeapon, BeamColloidalPhaser},
	"Gatling Gun":              {CategoryBeamWeapon, BeamGatlingGun},
	"Mini Blaster":             {CategoryBeamWeapon, BeamMiniBlaster},
	"Bludgeon":                 {CategoryBeamWeapon, BeamBludgeon},
	"Mark IV Blaster":          {CategoryBeamWeapon, BeamMarkIVBlaster},
	"Phased Sapper":            {CategoryBeamWeapon, BeamPhasedSapper},
	"Heavy Blaster":            {CategoryBeamWeapon, BeamHeavyBlaster},
	"Gatling Neutrino Cannon":  {CategoryBeamWeapon, BeamGatlingNeutrinoCannon},
	"Myopic Disruptor":         {CategoryBeamWeapon, BeamMyopicDisruptor},
	"Blunderbuss":              {CategoryBeamWeapon, BeamBlunderbuss},
	"Disruptor":                {CategoryBeamWeapon, BeamDisruptor},
	"Multi Contained Munition": {CategoryBeamWeapon, BeamMultiContainedMunition},
	"Syncro Sapper":            {CategoryBeamWeapon, BeamSyncroSapper},
	"Mega Disruptor":           {CategoryBeamWeapon, BeamMegaDisruptor},
	"Big Mutha Cannon":         {CategoryBeamWeapon, BeamBigMuthaCannon},
	"Streaming Pulverizer":     {CategoryBeamWeapon, BeamStreamingPulverizer},
	"Anti-Matter Pulverizer":   {CategoryBeamWeapon, BeamAntiMatterPulverizer},

	// Category 3: Torpedoes
	"Alpha Torpedo":       {CategoryTorpedo, TorpedoAlpha},
	"Beta Torpedo":        {CategoryTorpedo, TorpedoBeta},
	"Delta Torpedo":       {CategoryTorpedo, TorpedoDelta},
	"Epsilon Torpedo":     {CategoryTorpedo, TorpedoEpsilon},
	"Rho Torpedo":         {CategoryTorpedo, TorpedoRho},
	"Upsilon Torpedo":     {CategoryTorpedo, TorpedoUpsilon},
	"Omega Torpedo":       {CategoryTorpedo, TorpedoOmega},
	"Anti Matter Torpedo": {CategoryTorpedo, TorpedoAntiMatter},
	"Jihad Missile":       {CategoryTorpedo, TorpedoJihad},
	"Juggernaut Missile":  {CategoryTorpedo, TorpedoJuggernaut},
	"Doomsday Missile":    {CategoryTorpedo, TorpedoDoomsday},
	"Armageddon Missile":  {CategoryTorpedo, TorpedoArmageddon},

	// Category 4: Bombs
	"Lady Finger Bomb":      {CategoryBomb, 1},
	"Black Cat Bomb":        {CategoryBomb, 2},
	"M-70 Bomb":             {CategoryBomb, 3},
	"M-80 Bomb":             {CategoryBomb, 4},
	"Cherry Bomb":           {CategoryBomb, 5},
	"LBU-17 Bomb":           {CategoryBomb, 6},
	"LBU-32 Bomb":           {CategoryBomb, 7},
	"LBU-74 Bomb":           {CategoryBomb, 8},
	"Hush-a-Boom":           {CategoryBomb, 9},
	"Retro Bomb":            {CategoryBomb, 10},
	"Smart Bomb":            {CategoryBomb, 11},
	"Neutron Bomb":          {CategoryBomb, 12},
	"Enriched Neutron Bomb": {CategoryBomb, 13},
	"Peerless Bomb":         {CategoryBomb, 14},
	"Annihilator Bomb":      {CategoryBomb, 15},

	// Category 5: Terraforming
	"Total Terraform +3":      {CategoryTerraform, 1},
	"Total Terraform +5":      {CategoryTerraform, 2},
	"Total Terraform +7":      {CategoryTerraform, 3},
	"Total Terraform +10":     {CategoryTerraform, 4},
	"Total Terraform +15":     {CategoryTerraform, 5},
	"Total Terraform +20":     {CategoryTerraform, 6},
	"Total Terraform +25":     {CategoryTerraform, 7},
	"Total Terraform +30":     {CategoryTerraform, 8},
	"Gravity Terraform +3":    {CategoryTerraform, 9},
	"Gravity Terraform +7":    {CategoryTerraform, 10},
	"Gravity Terraform +11":   {CategoryTerraform, 11},
	"Gravity Terraform +15":   {CategoryTerraform, 12},
	"Temp Terraform +3":       {CategoryTerraform, 13},
	"Temp Terraform +7":       {CategoryTerraform, 14},
	"Temp Terraform +11":      {CategoryTerraform, 15},
	"Temp Terraform +15":      {CategoryTerraform, 16},
	"Radiation Terraform +3":  {CategoryTerraform, 17},
	"Radiation Terraform +7":  {CategoryTerraform, 18},
	"Radiation Terraform +11": {CategoryTerraform, 19},
	"Radiation Terraform +15": {CategoryTerraform, 20},

	// Category 6: Planetary Installations
	"Viewer 50":        {CategoryPlanetary, 1},
	"Viewer 90":        {CategoryPlanetary, 2},
	"Scoper 150":       {CategoryPlanetary, 3},
	"Scoper 220":       {CategoryPlanetary, 4},
	"Scoper 280":       {CategoryPlanetary, 5},
	"Snooper 320X":     {CategoryPlanetary, 6},
	"Snooper 400X":     {CategoryPlanetary, 7},
	"Snooper 500X":     {CategoryPlanetary, 8},
	"Snooper 620X":     {CategoryPlanetary, 9},
	"SDI":              {CategoryPlanetary, 10},
	"Missile Battery":  {CategoryPlanetary, 11},
	"Laser Battery":    {CategoryPlanetary, 12},
	"Planetary Shield": {CategoryPlanetary, 13},
	"Neutron Shield":   {CategoryPlanetary, 14},
	"Genesis Device":   {CategoryPlanetary, 15},

	// Category 7: Mining Robots
	"Robo-Midget Miner": {CategoryMiningRobo, MiningRoboMidget},
	"Robo-Mini-Miner":   {CategoryMiningRobo, MiningRoboMini},
	"Robo-Miner":        {CategoryMiningRobo, MiningRobo},
	"Robo-Maxi-Miner":   {CategoryMiningRobo, MiningRoboMaxi},
	"Robo-Super-Miner":  {CategoryMiningRobo, MiningRoboSuper},
	"Robo-Ultra-Miner":  {CategoryMiningRobo, MiningRoboUltra},
	"Alien Miner":       {CategoryMiningRobo, MiningAlien},
	"Orbital Adjuster":  {CategoryMiningRobo, MiningOrbitalAdj},

	// Category 8: Mine Layers
	"Mine Dispenser 40":   {CategoryMineLayer, 1},
	"Mine Dispenser 50":   {CategoryMineLayer, 2},
	"Mine Dispenser 80":   {CategoryMineLayer, 3},
	"Mine Dispenser 130":  {CategoryMineLayer, 4},
	"Heavy Dispenser 50":  {CategoryMineLayer, 5},
	"Heavy Dispenser 110": {CategoryMineLayer, 6},
	"Heavy Dispenser 200": {CategoryMineLayer, 7},
	"Speed Trap 20":       {CategoryMineLayer, 8},
	"Speed Trap 30":       {CategoryMineLayer, 9},
	"Speed Trap 50":       {CategoryMineLayer, 10},

	// Category 9: Mechanical
	"Colonization Module":         {CategoryMechanical, MechColonizationModule},
	"Orbital Construction Module": {CategoryMechanical, MechOrbitalConstructionModule},
	"Cargo Pod":                   {CategoryMechanical, MechCargoPod},
	"Super Cargo Pod":             {CategoryMechanical, MechSuperCargoPod},
	"Multi Cargo Pod":             {CategoryMechanical, MechMultiCargoPod},
	"Fuel Tank":                   {CategoryMechanical, MechFuelTank},
	"Super Fuel Tank":             {CategoryMechanical, MechSuperFuelTank},
	"Maneuvering Jet":             {CategoryMechanical, MechManeuveringJet},
	"Overthruster":                {CategoryMechanical, MechOverthruster},
	"Jump Gate":                   {CategoryMechanical, MechJumpGate},
	"Beam Deflector":              {CategoryMechanical, MechBeamDeflector},

	// Category 10: Electrical
	"Transport Cloaking":    {CategoryElectrical, 1},
	"Stealth Cloak":         {CategoryElectrical, 2},
	"Super-Stealth Cloak":   {CategoryElectrical, 3},
	"Ultra-Stealth Cloak":   {CategoryElectrical, 4},
	"Multi Function Pod":    {CategoryElectrical, 5},
	"Battle Computer":       {CategoryElectrical, 6},
	"Battle Super Computer": {CategoryElectrical, 7},
	"Battle Nexus":          {CategoryElectrical, 8},
	"Jammer 10":             {CategoryElectrical, 9},
	"Jammer 20":             {CategoryElectrical, 10},
	"Jammer 30":             {CategoryElectrical, 11},
	"Jammer 50":             {CategoryElectrical, 12},
	"Energy Capacitor":      {CategoryElectrical, 13},
	"Flux Capacitor":        {CategoryElectrical, 14},
	"Energy Dampener":       {CategoryElectrical, 15},
	"Tachyon Detector":      {CategoryElectrical, 16},
	"Anti-matter Generator": {CategoryElectrical, 17},

	// Category 11: Shields
	"Mole-skin Shield":         {CategoryShield, ShieldMoleskin},
	"Cow-hide Shield":          {CategoryShield, ShieldCowhide},
	"Wolverine Diffuse Shield": {CategoryShield, ShieldWolverineDiffuse},
	"Croby Sharmor":            {CategoryShield, ShieldCrobySharmor},
	"Shadow Shield":            {CategoryShield, ShieldShadow},
	"Bear Neutrino Barrier":    {CategoryShield, ShieldBearNeutrino},
	"Langston Shell":           {CategoryShield, ShieldLangstonShell},
	"Gorilla Delagator":        {CategoryShield, ShieldGorillaDelagator},
	"Elephant Hide Fortress":   {CategoryShield, ShieldElephantHide},
	"Complete Phase Shield":    {CategoryShield, ShieldCompletePhase},

	// Category 12: Scanners
	"Bat Scanner":          {CategoryScanner, ScannerBat},
	"Rhino Scanner":        {CategoryScanner, ScannerRhino},
	"Mole Scanner":         {CategoryScanner, ScannerMole},
	"DNA Scanner":          {CategoryScanner, ScannerDNA},
	"Possum Scanner":       {CategoryScanner, ScannerPossum},
	"Pick Pocket Scanner":  {CategoryScanner, ScannerPickPocket},
	"Chameleon Scanner":    {CategoryScanner, ScannerChameleon},
	"Ferret Scanner":       {CategoryScanner, ScannerFerret},
	"Dolphin Scanner":      {CategoryScanner, ScannerDolphin},
	"Gazelle Scanner":      {CategoryScanner, ScannerGazelle},
	"RNA Scanner":          {CategoryScanner, ScannerRNA},
	"Cheetah Scanner":      {CategoryScanner, ScannerCheetah},
	"Elephant Scanner":     {CategoryScanner, ScannerElephant},
	"Eagle Eye Scanner":    {CategoryScanner, ScannerEagleEye},
	"Robber Baron Scanner": {CategoryScanner, ScannerRobberBaron},
	"Peerless Scanner":     {CategoryScanner, ScannerPeerless},

	// Category 13: Armor
	"Tritanium":           {CategoryArmor, ArmorTritanium},
	"Crobmnium":           {CategoryArmor, ArmorCrobmnium},
	"Carbonic Armor":      {CategoryArmor, ArmorCarbonic},
	"Strobnium":           {CategoryArmor, ArmorStrobnium},
	"Organic Armor":       {CategoryArmor, ArmorOrganic},
	"Kelarium":            {CategoryArmor, ArmorKelarium},
	"Fielded Kelarium":    {CategoryArmor, ArmorFieldedKelarium},
	"Depleted Neutronium": {CategoryArmor, ArmorDepletedNeutronium},
	"Neutronium":          {CategoryArmor, ArmorNeutronium},
	"Mega Poly Shell":     {CategoryArmor, ArmorMegaPolyShell},
	"Valanium":            {CategoryArmor, ArmorValanium},
	"Superlatanium":       {CategoryArmor, ArmorSuperlatanium},

	// Category 14: Engines
	"Settler's Delight":          {CategoryEngine, EngineSettlersDelight},
	"Quick Jump 5":               {CategoryEngine, EngineQuickJump5},
	"Fuel Mizer":                 {CategoryEngine, EngineFuelMizer},
	"Long Hump 6":                {CategoryEngine, EngineLongHump6},
	"Daddy Long Legs 7":          {CategoryEngine, EngineDaddyLongLegs7},
	"Alpha Drive 8":              {CategoryEngine, EngineAlphaDrive8},
	"Trans-Galactic Drive":       {CategoryEngine, EngineTransGalacticDrive},
	"Interspace-10":              {CategoryEngine, EngineInterspace10},
	"Enigma Pulsar":              {CategoryEngine, EngineEnigmaPulsar},
	"Trans-Star 10":              {CategoryEngine, EngineTransStar10},
	"Radiating Hydro-Ram Scoop":  {CategoryEngine, EngineRadiatingHydroRamScoop},
	"Sub-Galactic Fuel Scoop":    {CategoryEngine, EngineSubGalacticFuelScoop},
	"Trans-Galactic Fuel Scoop":  {CategoryEngine, EngineTransGalacticFuelScoop},
	"Trans-Galactic Super Scoop": {CategoryEngine, EngineTransGalacticSuperScoop},
	"Trans-Galactic Mizer Scoop": {CategoryEngine, EngineTransGalacticMizerScoop},
	"Galaxy Scoop":               {CategoryEngine, EngineGalaxyScoop},
}

// ItemIDToName maps category and item ID to name.
var ItemIDToName = make(map[ItemCategory]map[int]string)

func init() {
	// Build reverse lookup maps
	for name, info := range ItemNameToInfo {
		if ItemIDToName[info.Category] == nil {
			ItemIDToName[info.Category] = make(map[int]string)
		}
		ItemIDToName[info.Category][info.ItemID] = name
	}
}

// GetItemName returns the name for a given category and item ID.
func GetItemName(category ItemCategory, itemID int) string {
	if categoryMap, ok := ItemIDToName[category]; ok {
		if name, ok := categoryMap[itemID]; ok {
			return name
		}
	}
	return ""
}

// GetItemInfo returns the category and item ID for a given name.
// Returns (ItemInfo{}, false) if not found.
func GetItemInfo(name string) (ItemInfo, bool) {
	info, ok := ItemNameToInfo[name]
	return info, ok
}

// TechRequirements represents the tech levels needed to build an item
type TechRequirements struct {
	Energy       int
	Weapons      int
	Propulsion   int
	Construction int
	Electronics  int
	Biotech      int
}

// CanBuildWith returns true if the given tech levels meet all requirements.
func (t TechRequirements) CanBuildWith(have TechRequirements) bool {
	return have.Energy >= t.Energy &&
		have.Weapons >= t.Weapons &&
		have.Propulsion >= t.Propulsion &&
		have.Construction >= t.Construction &&
		have.Electronics >= t.Electronics &&
		have.Biotech >= t.Biotech
}

// Cost represents the mineral and resource cost of an item
type Cost struct {
	Resources int
	Ironium   int
	Boranium  int
	Germanium int
}

// Engine represents an engine with its stats
type Engine struct {
	ID          int
	Name        string
	Tech        TechRequirements
	Mass        int
	Cost        Cost
	SafeSpeed   int     // Max warp without damage
	FreeSpeed   int     // Max warp that costs 0 fuel (for ramscoops)
	FuelPerMg   [11]int // Fuel consumption per 10mg at each warp (0-10)
	BattleSpeed int     // Speed bonus in battle (for Enigma Pulsar)
}

// Engines contains all engine definitions
var Engines = map[int]*Engine{
	EngineSettlersDelight: {
		ID: EngineSettlersDelight, Name: "Settler's Delight",
		Tech: TechRequirements{}, Mass: 2, Cost: Cost{2, 1, 0, 1},
		SafeSpeed: 6, FreeSpeed: 0,
		FuelPerMg: [11]int{0, 0, 0, 0, 0, 0, 0, 140, 275, 480, 576},
	},
	EngineQuickJump5: {
		ID: EngineQuickJump5, Name: "Quick Jump 5",
		Tech: TechRequirements{}, Mass: 4, Cost: Cost{3, 3, 0, 1},
		SafeSpeed: 5, FreeSpeed: 0,
		FuelPerMg: [11]int{0, 0, 25, 100, 100, 100, 180, 500, 800, 900, 1080},
	},
	EngineFuelMizer: {
		ID: EngineFuelMizer, Name: "Fuel Mizer",
		Tech: TechRequirements{Propulsion: 2}, Mass: 6, Cost: Cost{11, 8, 0, 0},
		SafeSpeed: 6, FreeSpeed: 0,
		FuelPerMg: [11]int{0, 0, 0, 0, 0, 35, 120, 175, 235, 360, 420},
	},
	EngineLongHump6: {
		ID: EngineLongHump6, Name: "Long Hump 6",
		Tech: TechRequirements{Propulsion: 3}, Mass: 9, Cost: Cost{6, 5, 0, 1},
		SafeSpeed: 6, FreeSpeed: 0,
		FuelPerMg: [11]int{0, 0, 20, 60, 100, 100, 105, 450, 750, 900, 1080},
	},
	EngineDaddyLongLegs7: {
		ID: EngineDaddyLongLegs7, Name: "Daddy Long Legs 7",
		Tech: TechRequirements{Propulsion: 5}, Mass: 13, Cost: Cost{12, 11, 0, 3},
		SafeSpeed: 7, FreeSpeed: 0,
		FuelPerMg: [11]int{0, 0, 20, 60, 70, 100, 100, 110, 600, 750, 900},
	},
	EngineAlphaDrive8: {
		ID: EngineAlphaDrive8, Name: "Alpha Drive 8",
		Tech: TechRequirements{Propulsion: 7}, Mass: 17, Cost: Cost{28, 16, 0, 3},
		SafeSpeed: 8, FreeSpeed: 0,
		FuelPerMg: [11]int{0, 0, 15, 50, 60, 70, 100, 100, 115, 700, 840},
	},
	EngineTransGalacticDrive: {
		ID: EngineTransGalacticDrive, Name: "Trans-Galactic Drive",
		Tech: TechRequirements{Propulsion: 9}, Mass: 25, Cost: Cost{50, 20, 20, 9},
		SafeSpeed: 9, FreeSpeed: 0,
		FuelPerMg: [11]int{0, 0, 15, 35, 45, 55, 70, 80, 90, 100, 120},
	},
	EngineInterspace10: {
		ID: EngineInterspace10, Name: "Interspace-10",
		Tech: TechRequirements{Propulsion: 11}, Mass: 25, Cost: Cost{60, 18, 25, 10},
		SafeSpeed: 10, FreeSpeed: 0,
		FuelPerMg: [11]int{0, 5, 10, 30, 40, 50, 60, 70, 80, 90, 100},
	},
	EngineEnigmaPulsar: {
		ID: EngineEnigmaPulsar, Name: "Enigma Pulsar",
		Tech: TechRequirements{Energy: 7, Propulsion: 13, Electronics: 5, Biotech: 9},
		Mass: 20, Cost: Cost{40, 12, 15, 11},
		SafeSpeed: 10, FreeSpeed: 0, BattleSpeed: 1,
		FuelPerMg: [11]int{0, 0, 0, 0, 0, 0, 0, 65, 75, 85, 95},
	},
	EngineTransStar10: {
		ID: EngineTransStar10, Name: "Trans-Star 10",
		Tech: TechRequirements{Propulsion: 23}, Mass: 5, Cost: Cost{10, 3, 0, 3},
		SafeSpeed: 10, FreeSpeed: 0,
		FuelPerMg: [11]int{0, 0, 5, 15, 20, 25, 30, 35, 40, 45, 50},
	},
	EngineRadiatingHydroRamScoop: {
		ID: EngineRadiatingHydroRamScoop, Name: "Radiating Hydro-Ram Scoop",
		Tech: TechRequirements{Energy: 2, Propulsion: 6}, Mass: 10, Cost: Cost{8, 3, 2, 9},
		SafeSpeed: 6, FreeSpeed: 6,
		FuelPerMg: [11]int{0, 0, 0, 0, 0, 0, 0, 165, 375, 600, 720},
	},
	EngineSubGalacticFuelScoop: {
		ID: EngineSubGalacticFuelScoop, Name: "Sub-Galactic Fuel Scoop",
		Tech: TechRequirements{Energy: 2, Propulsion: 8}, Mass: 20, Cost: Cost{12, 4, 4, 7},
		SafeSpeed: 7, FreeSpeed: 7,
		FuelPerMg: [11]int{0, 0, 0, 0, 0, 0, 0, 85, 105, 210, 380},
	},
	EngineTransGalacticFuelScoop: {
		ID: EngineTransGalacticFuelScoop, Name: "Trans-Galactic Fuel Scoop",
		Tech: TechRequirements{Energy: 3, Propulsion: 9}, Mass: 19, Cost: Cost{18, 5, 4, 12},
		SafeSpeed: 8, FreeSpeed: 8,
		FuelPerMg: [11]int{0, 0, 0, 0, 0, 0, 0, 0, 88, 100, 145},
	},
	EngineTransGalacticSuperScoop: {
		ID: EngineTransGalacticSuperScoop, Name: "Trans-Galactic Super Scoop",
		Tech: TechRequirements{Energy: 4, Propulsion: 12}, Mass: 18, Cost: Cost{24, 6, 4, 16},
		SafeSpeed: 9, FreeSpeed: 9,
		FuelPerMg: [11]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 65, 90},
	},
	EngineTransGalacticMizerScoop: {
		ID: EngineTransGalacticMizerScoop, Name: "Trans-Galactic Mizer Scoop",
		Tech: TechRequirements{Energy: 4, Propulsion: 16}, Mass: 11, Cost: Cost{20, 5, 2, 13},
		SafeSpeed: 10, FreeSpeed: 10,
		FuelPerMg: [11]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 70},
	},
	EngineGalaxyScoop: {
		ID: EngineGalaxyScoop, Name: "Galaxy Scoop",
		Tech: TechRequirements{Energy: 5, Propulsion: 20}, Mass: 8, Cost: Cost{12, 4, 2, 9},
		SafeSpeed: 10, FreeSpeed: 10,
		FuelPerMg: [11]int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	},
}

// Shield represents a shield with its stats
type Shield struct {
	ID           int
	Name         string
	Tech         TechRequirements
	Mass         int
	Cost         Cost
	ShieldValue  int
	ArmorValue   int // For Croby Sharmor and Langston Shell
	CloakPercent int // Cloaking percentage (for Shadow Shield)
}

// Shields contains all shield definitions
var Shields = map[int]*Shield{
	ShieldMoleskin: {
		ID: ShieldMoleskin, Name: "Mole-skin Shield",
		Tech: TechRequirements{}, Mass: 1, Cost: Cost{4, 1, 0, 1},
		ShieldValue: 25,
	},
	ShieldCowhide: {
		ID: ShieldCowhide, Name: "Cow-hide Shield",
		Tech: TechRequirements{Energy: 3}, Mass: 1, Cost: Cost{5, 2, 0, 2},
		ShieldValue: 40,
	},
	ShieldWolverineDiffuse: {
		ID: ShieldWolverineDiffuse, Name: "Wolverine Diffuse Shield",
		Tech: TechRequirements{Energy: 6}, Mass: 1, Cost: Cost{6, 3, 0, 3},
		ShieldValue: 60,
	},
	ShieldCrobySharmor: {
		ID: ShieldCrobySharmor, Name: "Croby Sharmor",
		Tech: TechRequirements{Energy: 7, Construction: 4}, Mass: 10, Cost: Cost{15, 7, 0, 4},
		ShieldValue: 60, ArmorValue: 65,
	},
	ShieldShadow: {
		ID: ShieldShadow, Name: "Shadow Shield",
		Tech: TechRequirements{Energy: 7, Electronics: 3}, Mass: 2, Cost: Cost{7, 3, 0, 3},
		ShieldValue: 75, CloakPercent: 70,
	},
	ShieldBearNeutrino: {
		ID: ShieldBearNeutrino, Name: "Bear Neutrino Barrier",
		Tech: TechRequirements{Energy: 10}, Mass: 1, Cost: Cost{8, 4, 0, 4},
		ShieldValue: 100,
	},
	ShieldLangstonShell: {
		ID: ShieldLangstonShell, Name: "Langston Shell",
		Tech: TechRequirements{Energy: 12, Propulsion: 9, Electronics: 9}, Mass: 10, Cost: Cost{20, 10, 2, 6},
		ShieldValue: 125, ArmorValue: 65,
	},
	ShieldGorillaDelagator: {
		ID: ShieldGorillaDelagator, Name: "Gorilla Delagator",
		Tech: TechRequirements{Energy: 14}, Mass: 1, Cost: Cost{11, 5, 0, 6},
		ShieldValue: 175,
	},
	ShieldElephantHide: {
		ID: ShieldElephantHide, Name: "Elephant Hide Fortress",
		Tech: TechRequirements{Energy: 18}, Mass: 1, Cost: Cost{15, 8, 0, 10},
		ShieldValue: 300,
	},
	ShieldCompletePhase: {
		ID: ShieldCompletePhase, Name: "Complete Phase Shield",
		Tech: TechRequirements{Energy: 22}, Mass: 1, Cost: Cost{20, 12, 0, 15},
		ShieldValue: 500,
	},
}

// Armor represents armor with its stats
type Armor struct {
	ID           int
	Name         string
	Tech         TechRequirements
	Mass         int
	Cost         Cost
	ArmorValue   int
	CloakPercent int // Cloaking percentage (for Depleted Neutronium)
}

// Armors contains all armor definitions
var Armors = map[int]*Armor{
	ArmorTritanium: {
		ID: ArmorTritanium, Name: "Tritanium",
		Tech: TechRequirements{}, Mass: 60, Cost: Cost{10, 5, 0, 0},
		ArmorValue: 50,
	},
	ArmorCrobmnium: {
		ID: ArmorCrobmnium, Name: "Crobmnium",
		Tech: TechRequirements{Construction: 3}, Mass: 56, Cost: Cost{13, 6, 0, 0},
		ArmorValue: 75,
	},
	ArmorCarbonic: {
		ID: ArmorCarbonic, Name: "Carbonic Armor",
		Tech: TechRequirements{Biotech: 4}, Mass: 25, Cost: Cost{15, 0, 0, 5},
		ArmorValue: 100,
	},
	ArmorStrobnium: {
		ID: ArmorStrobnium, Name: "Strobnium",
		Tech: TechRequirements{Construction: 6}, Mass: 54, Cost: Cost{18, 8, 0, 0},
		ArmorValue: 120,
	},
	ArmorOrganic: {
		ID: ArmorOrganic, Name: "Organic Armor",
		Tech: TechRequirements{Biotech: 7}, Mass: 15, Cost: Cost{20, 0, 0, 6},
		ArmorValue: 175,
	},
	ArmorKelarium: {
		ID: ArmorKelarium, Name: "Kelarium",
		Tech: TechRequirements{Construction: 9}, Mass: 50, Cost: Cost{25, 9, 1, 0},
		ArmorValue: 180,
	},
	ArmorFieldedKelarium: {
		ID: ArmorFieldedKelarium, Name: "Fielded Kelarium",
		Tech: TechRequirements{Energy: 4, Construction: 10}, Mass: 50, Cost: Cost{28, 10, 0, 2},
		ArmorValue: 175,
	},
	ArmorDepletedNeutronium: {
		ID: ArmorDepletedNeutronium, Name: "Depleted Neutronium",
		Tech: TechRequirements{Construction: 10, Electronics: 3}, Mass: 50, Cost: Cost{28, 10, 0, 2},
		ArmorValue: 200, CloakPercent: 50,
	},
	ArmorNeutronium: {
		ID: ArmorNeutronium, Name: "Neutronium",
		Tech: TechRequirements{Construction: 12}, Mass: 45, Cost: Cost{30, 11, 2, 1},
		ArmorValue: 275,
	},
	ArmorMegaPolyShell: {
		ID: ArmorMegaPolyShell, Name: "Mega Poly Shell",
		Tech: TechRequirements{Energy: 14, Construction: 14, Electronics: 14, Biotech: 6},
		Mass: 20, Cost: Cost{65, 18, 6, 6},
		ArmorValue: 400,
	},
	ArmorValanium: {
		ID: ArmorValanium, Name: "Valanium",
		Tech: TechRequirements{Construction: 16}, Mass: 40, Cost: Cost{50, 15, 0, 0},
		ArmorValue: 500,
	},
	ArmorSuperlatanium: {
		ID: ArmorSuperlatanium, Name: "Superlatanium",
		Tech: TechRequirements{Construction: 24}, Mass: 30, Cost: Cost{100, 25, 0, 0},
		ArmorValue: 1500,
	},
}

// BeamWeapon represents a beam weapon with its stats
type BeamWeapon struct {
	ID         int
	Name       string
	Tech       TechRequirements
	Mass       int
	Cost       Cost
	Power      int  // Damage at range 0
	Range      int  // Max range in squares
	Initiative int  // Higher fires first
	IsSapper   bool // Drains shields
	IsGatling  bool // Hits all targets
}

// BeamWeapons contains all beam weapon definitions
var BeamWeapons = map[int]*BeamWeapon{
	BeamLaser:                  {ID: BeamLaser, Name: "Laser", Tech: TechRequirements{}, Mass: 1, Cost: Cost{5, 0, 6, 0}, Power: 10, Range: 1, Initiative: 9},
	BeamXRayLaser:              {ID: BeamXRayLaser, Name: "X-Ray Laser", Tech: TechRequirements{Weapons: 3}, Mass: 1, Cost: Cost{6, 0, 6, 0}, Power: 16, Range: 1, Initiative: 9},
	BeamMiniGun:                {ID: BeamMiniGun, Name: "Mini Gun", Tech: TechRequirements{Weapons: 5}, Mass: 3, Cost: Cost{10, 0, 16, 0}, Power: 13, Range: 2, Initiative: 12, IsGatling: true},
	BeamYakimoraLightPhaser:    {ID: BeamYakimoraLightPhaser, Name: "Yakimora Light Phaser", Tech: TechRequirements{Weapons: 6}, Mass: 1, Cost: Cost{7, 0, 8, 0}, Power: 26, Range: 1, Initiative: 9},
	BeamBlackjack:              {ID: BeamBlackjack, Name: "Blackjack", Tech: TechRequirements{Weapons: 7}, Mass: 10, Cost: Cost{7, 0, 16, 0}, Power: 90, Range: 0, Initiative: 10},
	BeamPhaserBazooka:          {ID: BeamPhaserBazooka, Name: "Phaser Bazooka", Tech: TechRequirements{Weapons: 8}, Mass: 2, Cost: Cost{11, 0, 8, 0}, Power: 26, Range: 2, Initiative: 7},
	BeamPulsedSapper:           {ID: BeamPulsedSapper, Name: "Pulsed Sapper", Tech: TechRequirements{Energy: 5, Weapons: 9}, Mass: 1, Cost: Cost{12, 0, 0, 4}, Power: 82, Range: 3, Initiative: 14, IsSapper: true},
	BeamColloidalPhaser:        {ID: BeamColloidalPhaser, Name: "Colloidal Phaser", Tech: TechRequirements{Weapons: 10}, Mass: 2, Cost: Cost{18, 0, 14, 0}, Power: 26, Range: 3, Initiative: 5},
	BeamGatlingGun:             {ID: BeamGatlingGun, Name: "Gatling Gun", Tech: TechRequirements{Weapons: 11}, Mass: 3, Cost: Cost{13, 0, 20, 0}, Power: 31, Range: 2, Initiative: 12, IsGatling: true},
	BeamMiniBlaster:            {ID: BeamMiniBlaster, Name: "Mini Blaster", Tech: TechRequirements{Weapons: 12}, Mass: 1, Cost: Cost{9, 0, 10, 0}, Power: 66, Range: 1, Initiative: 9},
	BeamBludgeon:               {ID: BeamBludgeon, Name: "Bludgeon", Tech: TechRequirements{Weapons: 13}, Mass: 10, Cost: Cost{9, 0, 22, 0}, Power: 231, Range: 0, Initiative: 10},
	BeamMarkIVBlaster:          {ID: BeamMarkIVBlaster, Name: "Mark IV Blaster", Tech: TechRequirements{Weapons: 14}, Mass: 2, Cost: Cost{15, 0, 12, 0}, Power: 66, Range: 2, Initiative: 7},
	BeamPhasedSapper:           {ID: BeamPhasedSapper, Name: "Phased Sapper", Tech: TechRequirements{Energy: 8, Weapons: 15}, Mass: 1, Cost: Cost{16, 0, 0, 6}, Power: 211, Range: 3, Initiative: 14, IsSapper: true},
	BeamHeavyBlaster:           {ID: BeamHeavyBlaster, Name: "Heavy Blaster", Tech: TechRequirements{Weapons: 16}, Mass: 2, Cost: Cost{25, 0, 20, 0}, Power: 66, Range: 3, Initiative: 5},
	BeamGatlingNeutrinoCannon:  {ID: BeamGatlingNeutrinoCannon, Name: "Gatling Neutrino Cannon", Tech: TechRequirements{Weapons: 17}, Mass: 3, Cost: Cost{17, 0, 28, 0}, Power: 80, Range: 2, Initiative: 13, IsGatling: true},
	BeamMyopicDisruptor:        {ID: BeamMyopicDisruptor, Name: "Myopic Disruptor", Tech: TechRequirements{Weapons: 18}, Mass: 1, Cost: Cost{12, 0, 14, 0}, Power: 169, Range: 1, Initiative: 9},
	BeamBlunderbuss:            {ID: BeamBlunderbuss, Name: "Blunderbuss", Tech: TechRequirements{Weapons: 19}, Mass: 10, Cost: Cost{13, 0, 30, 0}, Power: 592, Range: 0, Initiative: 11},
	BeamDisruptor:              {ID: BeamDisruptor, Name: "Disruptor", Tech: TechRequirements{Weapons: 20}, Mass: 2, Cost: Cost{20, 0, 16, 0}, Power: 169, Range: 2, Initiative: 8},
	BeamMultiContainedMunition: {ID: BeamMultiContainedMunition, Name: "Multi Contained Munition", Tech: TechRequirements{Energy: 21, Weapons: 21, Electronics: 16, Biotech: 12}, Mass: 8, Cost: Cost{40, 6, 40, 6}, Power: 140, Range: 3, Initiative: 6},
	BeamSyncroSapper:           {ID: BeamSyncroSapper, Name: "Syncro Sapper", Tech: TechRequirements{Energy: 11, Weapons: 21}, Mass: 1, Cost: Cost{21, 0, 0, 8}, Power: 541, Range: 3, Initiative: 14, IsSapper: true},
	BeamMegaDisruptor:          {ID: BeamMegaDisruptor, Name: "Mega Disruptor", Tech: TechRequirements{Weapons: 22}, Mass: 2, Cost: Cost{33, 0, 30, 0}, Power: 169, Range: 3, Initiative: 6},
	BeamBigMuthaCannon:         {ID: BeamBigMuthaCannon, Name: "Big Mutha Cannon", Tech: TechRequirements{Weapons: 23}, Mass: 3, Cost: Cost{23, 0, 36, 0}, Power: 204, Range: 2, Initiative: 13, IsGatling: true},
	BeamStreamingPulverizer:    {ID: BeamStreamingPulverizer, Name: "Streaming Pulverizer", Tech: TechRequirements{Weapons: 24}, Mass: 1, Cost: Cost{16, 0, 20, 0}, Power: 433, Range: 1, Initiative: 9},
	BeamAntiMatterPulverizer:   {ID: BeamAntiMatterPulverizer, Name: "Anti-Matter Pulverizer", Tech: TechRequirements{Weapons: 26}, Mass: 2, Cost: Cost{27, 0, 22, 0}, Power: 433, Range: 2, Initiative: 8},
}

// Torpedo represents a torpedo/missile with its stats
type Torpedo struct {
	ID         int
	Name       string
	Tech       TechRequirements
	Mass       int
	Cost       Cost
	Power      int // Damage per hit
	Range      int // Initiative/range
	Accuracy   int // Base accuracy percentage
	Initiative int
	IsCapital  bool // Capital missiles (Jihad, Juggernaut, Doomsday, Armageddon)
}

// Torpedoes contains all torpedo definitions
var Torpedoes = map[int]*Torpedo{
	TorpedoAlpha:      {ID: TorpedoAlpha, Name: "Alpha Torpedo", Tech: TechRequirements{}, Mass: 25, Cost: Cost{5, 9, 3, 3}, Power: 5, Range: 4, Accuracy: 35},
	TorpedoBeta:       {ID: TorpedoBeta, Name: "Beta Torpedo", Tech: TechRequirements{Weapons: 5, Propulsion: 1}, Mass: 25, Cost: Cost{6, 18, 6, 4}, Power: 12, Range: 4, Accuracy: 45, Initiative: 1},
	TorpedoDelta:      {ID: TorpedoDelta, Name: "Delta Torpedo", Tech: TechRequirements{Weapons: 10, Propulsion: 2}, Mass: 25, Cost: Cost{8, 22, 8, 5}, Power: 26, Range: 4, Accuracy: 60, Initiative: 1},
	TorpedoEpsilon:    {ID: TorpedoEpsilon, Name: "Epsilon Torpedo", Tech: TechRequirements{Weapons: 14, Propulsion: 3}, Mass: 25, Cost: Cost{10, 30, 10, 6}, Power: 48, Range: 5, Accuracy: 65, Initiative: 2},
	TorpedoRho:        {ID: TorpedoRho, Name: "Rho Torpedo", Tech: TechRequirements{Weapons: 18, Propulsion: 4}, Mass: 25, Cost: Cost{12, 34, 12, 8}, Power: 90, Range: 5, Accuracy: 75, Initiative: 2},
	TorpedoUpsilon:    {ID: TorpedoUpsilon, Name: "Upsilon Torpedo", Tech: TechRequirements{Weapons: 22, Propulsion: 5}, Mass: 25, Cost: Cost{15, 40, 14, 9}, Power: 169, Range: 5, Accuracy: 75, Initiative: 3},
	TorpedoOmega:      {ID: TorpedoOmega, Name: "Omega Torpedo", Tech: TechRequirements{Weapons: 26, Propulsion: 6}, Mass: 25, Cost: Cost{18, 52, 18, 12}, Power: 316, Range: 5, Accuracy: 80, Initiative: 4},
	TorpedoAntiMatter: {ID: TorpedoAntiMatter, Name: "Anti Matter Torpedo", Tech: TechRequirements{Weapons: 11, Propulsion: 12, Biotech: 21}, Mass: 8, Cost: Cost{50, 3, 8, 1}, Power: 60, Range: 6, Accuracy: 85},
	TorpedoJihad:      {ID: TorpedoJihad, Name: "Jihad Missile", Tech: TechRequirements{Weapons: 12, Propulsion: 6}, Mass: 35, Cost: Cost{13, 37, 13, 9}, Power: 85, Range: 5, Accuracy: 20, IsCapital: true},
	TorpedoJuggernaut: {ID: TorpedoJuggernaut, Name: "Juggernaut Missile", Tech: TechRequirements{Weapons: 16, Propulsion: 8}, Mass: 35, Cost: Cost{16, 48, 16, 11}, Power: 150, Range: 5, Accuracy: 20, Initiative: 1, IsCapital: true},
	TorpedoDoomsday:   {ID: TorpedoDoomsday, Name: "Doomsday Missile", Tech: TechRequirements{Weapons: 20, Propulsion: 10}, Mass: 35, Cost: Cost{20, 60, 20, 13}, Power: 280, Range: 6, Accuracy: 25, Initiative: 2, IsCapital: true},
	TorpedoArmageddon: {ID: TorpedoArmageddon, Name: "Armageddon Missile", Tech: TechRequirements{Weapons: 24, Propulsion: 10}, Mass: 35, Cost: Cost{24, 67, 23, 16}, Power: 525, Range: 6, Accuracy: 30, Initiative: 3, IsCapital: true},
}

// Electrical IDs (1-indexed)
// NOTE: Decompiled SpdOfShip() references "Thruster" at item 4 in category 0x800 (Electrical).
// This conflicts with ElecUltraStealthCloak. The slot may have dual-purpose items,
// or the decompiled code uses different item numbering. Need test data to validate.
const (
	ElecTransportCloaking = iota + 1
	ElecStealthCloak
	ElecSuperStealthCloak
	ElecUltraStealthCloak       // Also ElecThruster per SpdOfShip() - +1 speed bonus
	ElecThruster            = 4 // Per decompiled SpdOfShip() - same slot as UltraStealthCloak
	ElecMultiFunctionPod    = 5
	ElecBattleComputer      = 6
	ElecBattleSuperComputer = 7
	ElecBattleNexus         = 8
	ElecJammer10            = 9
	ElecJammer20            = 10
	ElecJammer30            = 11
	ElecJammer50            = 12
	ElecEnergyCapacitor     = 13
	ElecFluxCapacitor       = 14
	ElecEnergyDampener      = 15
	ElecTachyonDetector     = 16
	ElecAntiMatterGenerator = 17
)

// Electrical represents an electrical component
type Electrical struct {
	ID               int
	Name             string
	Tech             TechRequirements
	Mass             int
	Cost             Cost
	CloakPercent     int // Cloaking percentage (0 = not a cloak)
	InitiativeBonus  int // For battle computers
	TorpedoAccuracy  int // Bonus to torpedo accuracy
	BeamDeflection   int // For jammers - reduces enemy beam accuracy
	CapacitorBonus   int // Extra beam damage percentage
	EnergyDampening  int // Reduces enemy initiative
	TachyonDetection int // Range to detect cloaked ships
	FuelGeneration   int // Fuel generated per year
}

// Electricals contains all electrical component definitions
var Electricals = map[int]*Electrical{
	ElecTransportCloaking:   {ID: ElecTransportCloaking, Name: "Transport Cloaking", Tech: TechRequirements{}, Mass: 1, Cost: Cost{3, 2, 0, 2}, CloakPercent: 300},
	ElecStealthCloak:        {ID: ElecStealthCloak, Name: "Stealth Cloak", Tech: TechRequirements{Energy: 2, Electronics: 5}, Mass: 2, Cost: Cost{5, 2, 0, 2}, CloakPercent: 70},
	ElecSuperStealthCloak:   {ID: ElecSuperStealthCloak, Name: "Super-Stealth Cloak", Tech: TechRequirements{Energy: 4, Electronics: 10}, Mass: 3, Cost: Cost{15, 8, 0, 8}, CloakPercent: 140},
	ElecUltraStealthCloak:   {ID: ElecUltraStealthCloak, Name: "Ultra-Stealth Cloak", Tech: TechRequirements{Energy: 10, Electronics: 12}, Mass: 5, Cost: Cost{25, 10, 0, 10}, CloakPercent: 540},
	ElecMultiFunctionPod:    {ID: ElecMultiFunctionPod, Name: "Multi Function Pod", Tech: TechRequirements{Energy: 11, Propulsion: 11, Electronics: 11}, Mass: 2, Cost: Cost{15, 5, 0, 5}, CloakPercent: 60},
	ElecBattleComputer:      {ID: ElecBattleComputer, Name: "Battle Computer", Tech: TechRequirements{}, Mass: 1, Cost: Cost{6, 0, 0, 15}, InitiativeBonus: 1, TorpedoAccuracy: 20},
	ElecBattleSuperComputer: {ID: ElecBattleSuperComputer, Name: "Battle Super Computer", Tech: TechRequirements{Energy: 5, Electronics: 11}, Mass: 1, Cost: Cost{14, 0, 0, 25}, InitiativeBonus: 2, TorpedoAccuracy: 30},
	ElecBattleNexus:         {ID: ElecBattleNexus, Name: "Battle Nexus", Tech: TechRequirements{Energy: 10, Electronics: 19}, Mass: 1, Cost: Cost{15, 0, 0, 30}, InitiativeBonus: 3, TorpedoAccuracy: 50},
	ElecJammer10:            {ID: ElecJammer10, Name: "Jammer 10", Tech: TechRequirements{Energy: 2, Electronics: 6}, Mass: 1, Cost: Cost{6, 0, 0, 2}, BeamDeflection: 10},
	ElecJammer20:            {ID: ElecJammer20, Name: "Jammer 20", Tech: TechRequirements{Energy: 4, Electronics: 10}, Mass: 1, Cost: Cost{20, 1, 0, 5}, BeamDeflection: 20},
	ElecJammer30:            {ID: ElecJammer30, Name: "Jammer 30", Tech: TechRequirements{Energy: 8, Electronics: 16}, Mass: 1, Cost: Cost{20, 1, 0, 6}, BeamDeflection: 30},
	ElecJammer50:            {ID: ElecJammer50, Name: "Jammer 50", Tech: TechRequirements{Energy: 16, Electronics: 22}, Mass: 1, Cost: Cost{20, 2, 0, 7}, BeamDeflection: 50},
	ElecEnergyCapacitor:     {ID: ElecEnergyCapacitor, Name: "Energy Capacitor", Tech: TechRequirements{Energy: 7, Electronics: 4}, Mass: 1, Cost: Cost{5, 0, 0, 8}, CapacitorBonus: 10},
	ElecFluxCapacitor:       {ID: ElecFluxCapacitor, Name: "Flux Capacitor", Tech: TechRequirements{Energy: 14, Electronics: 8}, Mass: 1, Cost: Cost{5, 0, 0, 8}, CapacitorBonus: 20},
	ElecEnergyDampener:      {ID: ElecEnergyDampener, Name: "Energy Dampener", Tech: TechRequirements{Energy: 14, Propulsion: 8}, Mass: 2, Cost: Cost{50, 5, 10, 0}, EnergyDampening: 1},
	ElecTachyonDetector:     {ID: ElecTachyonDetector, Name: "Tachyon Detector", Tech: TechRequirements{Energy: 8, Electronics: 14}, Mass: 1, Cost: Cost{70, 1, 5, 0}, TachyonDetection: 2},
	ElecAntiMatterGenerator: {ID: ElecAntiMatterGenerator, Name: "Anti-matter Generator", Tech: TechRequirements{Weapons: 12, Biotech: 7}, Mass: 10, Cost: Cost{10, 8, 3, 3}, FuelGeneration: 200},
}

// Mechanical represents a mechanical component
type Mechanical struct {
	ID            int
	Name          string
	Tech          TechRequirements
	Mass          int
	Cost          Cost
	CargoCapacity int  // Extra cargo space
	FuelCapacity  int  // Extra fuel capacity
	Colonizer     bool // Can colonize planets
	OrbitalBuild  bool // Can build starbases
	SpeedBonus    int  // Battle speed bonus
	BeamDeflect   int  // Beam deflection percentage
	ArmorValue    int  // For Multi Cargo Pod
}

// Mechanicals contains all mechanical component definitions
// NOTE: IDs reordered to match decompiled SpdOfShip() - Maneuver Jet=7, Overthruster=8
var Mechanicals = map[int]*Mechanical{
	MechColonizationModule:        {ID: MechColonizationModule, Name: "Colonization Module", Tech: TechRequirements{}, Mass: 32, Cost: Cost{10, 12, 10, 10}, Colonizer: true},
	MechOrbitalConstructionModule: {ID: MechOrbitalConstructionModule, Name: "Orbital Construction Module", Tech: TechRequirements{}, Mass: 50, Cost: Cost{20, 20, 15, 15}, OrbitalBuild: true},
	MechCargoPod:                  {ID: MechCargoPod, Name: "Cargo Pod", Tech: TechRequirements{Construction: 3}, Mass: 5, Cost: Cost{10, 5, 0, 2}, CargoCapacity: 50},
	MechSuperCargoPod:             {ID: MechSuperCargoPod, Name: "Super Cargo Pod", Tech: TechRequirements{Energy: 3, Construction: 9}, Mass: 7, Cost: Cost{15, 8, 0, 2}, CargoCapacity: 100},
	MechMultiCargoPod:             {ID: MechMultiCargoPod, Name: "Multi Cargo Pod", Tech: TechRequirements{Energy: 5, Construction: 11, Electronics: 5}, Mass: 9, Cost: Cost{25, 12, 0, 3}, CargoCapacity: 250, ArmorValue: 50},
	MechFuelTank:                  {ID: MechFuelTank, Name: "Fuel Tank", Tech: TechRequirements{}, Mass: 3, Cost: Cost{4, 6, 0, 0}, FuelCapacity: 250},
	MechManeuveringJet:            {ID: MechManeuveringJet, Name: "Maneuvering Jet", Tech: TechRequirements{Energy: 2, Propulsion: 3}, Mass: 5, Cost: Cost{10, 5, 0, 5}, SpeedBonus: 1},
	MechOverthruster:              {ID: MechOverthruster, Name: "Overthruster", Tech: TechRequirements{Energy: 5, Propulsion: 12}, Mass: 5, Cost: Cost{20, 10, 0, 8}, SpeedBonus: 2},
	MechSuperFuelTank:             {ID: MechSuperFuelTank, Name: "Super Fuel Tank", Tech: TechRequirements{Energy: 6, Propulsion: 4, Construction: 14}, Mass: 8, Cost: Cost{8, 8, 0, 0}, FuelCapacity: 500},
	MechJumpGate:                  {ID: MechJumpGate, Name: "Jump Gate", Tech: TechRequirements{Energy: 16, Propulsion: 20, Construction: 20, Electronics: 16}, Mass: 10, Cost: Cost{40, 0, 0, 50}},
	MechBeamDeflector:             {ID: MechBeamDeflector, Name: "Beam Deflector", Tech: TechRequirements{Energy: 6, Weapons: 6, Construction: 6, Electronics: 6}, Mass: 1, Cost: Cost{8, 0, 0, 10}, BeamDeflect: 10},
}

// MineLayer represents a mine layer
type MineLayer struct {
	ID           int
	Name         string
	Tech         TechRequirements
	Mass         int
	Cost         Cost
	MinesPerYear int
	MineType     string // "Normal", "Heavy", or "Speed"
}

// MineLayers contains all mine layer definitions
var MineLayers = map[int]*MineLayer{
	1:  {ID: 1, Name: "Mine Dispenser 40", Tech: TechRequirements{}, Mass: 25, Cost: Cost{45, 2, 10, 8}, MinesPerYear: 40, MineType: "Normal"},
	2:  {ID: 2, Name: "Mine Dispenser 50", Tech: TechRequirements{Energy: 2, Biotech: 4}, Mass: 30, Cost: Cost{55, 2, 12, 10}, MinesPerYear: 50, MineType: "Normal"},
	3:  {ID: 3, Name: "Mine Dispenser 80", Tech: TechRequirements{Energy: 3, Biotech: 7}, Mass: 30, Cost: Cost{65, 2, 14, 10}, MinesPerYear: 80, MineType: "Normal"},
	4:  {ID: 4, Name: "Mine Dispenser 130", Tech: TechRequirements{Energy: 6, Biotech: 12}, Mass: 30, Cost: Cost{80, 2, 18, 10}, MinesPerYear: 130, MineType: "Normal"},
	5:  {ID: 5, Name: "Heavy Dispenser 50", Tech: TechRequirements{Energy: 5, Biotech: 3}, Mass: 10, Cost: Cost{50, 2, 20, 5}, MinesPerYear: 50, MineType: "Heavy"},
	6:  {ID: 6, Name: "Heavy Dispenser 110", Tech: TechRequirements{Energy: 9, Biotech: 5}, Mass: 15, Cost: Cost{70, 2, 30, 5}, MinesPerYear: 110, MineType: "Heavy"},
	7:  {ID: 7, Name: "Heavy Dispenser 200", Tech: TechRequirements{Energy: 14, Biotech: 7}, Mass: 20, Cost: Cost{90, 2, 45, 5}, MinesPerYear: 200, MineType: "Heavy"},
	8:  {ID: 8, Name: "Speed Trap 20", Tech: TechRequirements{Propulsion: 2, Biotech: 2}, Mass: 100, Cost: Cost{60, 30, 0, 12}, MinesPerYear: 20, MineType: "Speed"},
	9:  {ID: 9, Name: "Speed Trap 30", Tech: TechRequirements{Propulsion: 3, Biotech: 6}, Mass: 135, Cost: Cost{72, 32, 0, 14}, MinesPerYear: 30, MineType: "Speed"},
	10: {ID: 10, Name: "Speed Trap 50", Tech: TechRequirements{Propulsion: 5, Biotech: 11}, Mass: 140, Cost: Cost{80, 40, 0, 15}, MinesPerYear: 50, MineType: "Speed"},
}

// MiningRobot represents a mining robot
type MiningRobot struct {
	ID            int
	Name          string
	Tech          TechRequirements
	Mass          int
	Cost          Cost
	MiningRate    int // kT per year
	TerraformRate int // For Orbital Adjuster
}

// MiningRobots contains all mining robot definitions
var MiningRobots = map[int]*MiningRobot{
	MiningRoboMidget: {ID: MiningRoboMidget, Name: "Robo-Midget Miner", Tech: TechRequirements{}, Mass: 80, Cost: Cost{50, 14, 0, 4}, MiningRate: 5},
	MiningRoboMini:   {ID: MiningRoboMini, Name: "Robo-Mini-Miner", Tech: TechRequirements{Construction: 2, Electronics: 1}, Mass: 240, Cost: Cost{100, 30, 0, 7}, MiningRate: 4},
	MiningRobo:       {ID: MiningRobo, Name: "Robo-Miner", Tech: TechRequirements{Construction: 4, Electronics: 2}, Mass: 240, Cost: Cost{100, 30, 0, 7}, MiningRate: 12},
	MiningRoboMaxi:   {ID: MiningRoboMaxi, Name: "Robo-Maxi-Miner", Tech: TechRequirements{Construction: 7, Electronics: 4}, Mass: 240, Cost: Cost{100, 30, 0, 7}, MiningRate: 18},
	MiningRoboSuper:  {ID: MiningRoboSuper, Name: "Robo-Super-Miner", Tech: TechRequirements{Construction: 12, Electronics: 6}, Mass: 240, Cost: Cost{100, 30, 0, 7}, MiningRate: 27},
	MiningRoboUltra:  {ID: MiningRoboUltra, Name: "Robo-Ultra-Miner", Tech: TechRequirements{Construction: 15, Electronics: 8}, Mass: 80, Cost: Cost{50, 14, 0, 4}, MiningRate: 25},
	MiningAlien:      {ID: MiningAlien, Name: "Alien Miner", Tech: TechRequirements{Energy: 5, Construction: 10, Electronics: 5, Biotech: 5}, Mass: 20, Cost: Cost{20, 8, 0, 2}, MiningRate: 10},
	MiningOrbitalAdj: {ID: MiningOrbitalAdj, Name: "Orbital Adjuster", Tech: TechRequirements{Biotech: 6}, Mass: 80, Cost: Cost{50, 25, 25, 25}, TerraformRate: 1},
}

// Bomb represents a bomb with its stats
type Bomb struct {
	ID              int
	Name            string
	Tech            TechRequirements
	Mass            int
	Cost            Cost
	KillRate        int  // Colonists killed per bomb (min %)
	StructureKill   int  // Structures destroyed per bomb
	IsSmart         bool // Smart bombs - bypass defenses
	UnterraformRate int  // For Retro Bomb
}

// Bombs contains all bomb definitions
var Bombs = map[int]*Bomb{
	1:  {ID: 1, Name: "Lady Finger Bomb", Tech: TechRequirements{Weapons: 2}, Mass: 40, Cost: Cost{5, 1, 20, 0}, KillRate: 6, StructureKill: 2},
	2:  {ID: 2, Name: "Black Cat Bomb", Tech: TechRequirements{Weapons: 5}, Mass: 45, Cost: Cost{7, 1, 22, 0}, KillRate: 9, StructureKill: 4},
	3:  {ID: 3, Name: "M-70 Bomb", Tech: TechRequirements{Weapons: 8}, Mass: 50, Cost: Cost{9, 1, 24, 0}, KillRate: 12, StructureKill: 6},
	4:  {ID: 4, Name: "M-80 Bomb", Tech: TechRequirements{Weapons: 11}, Mass: 55, Cost: Cost{12, 1, 25, 0}, KillRate: 17, StructureKill: 7},
	5:  {ID: 5, Name: "Cherry Bomb", Tech: TechRequirements{Weapons: 14}, Mass: 52, Cost: Cost{11, 1, 25, 0}, KillRate: 25, StructureKill: 10},
	6:  {ID: 6, Name: "LBU-17 Bomb", Tech: TechRequirements{Weapons: 5, Electronics: 8}, Mass: 30, Cost: Cost{7, 1, 15, 15}, KillRate: 2, StructureKill: 16},
	7:  {ID: 7, Name: "LBU-32 Bomb", Tech: TechRequirements{Weapons: 10, Electronics: 10}, Mass: 35, Cost: Cost{10, 1, 24, 15}, KillRate: 3, StructureKill: 28},
	8:  {ID: 8, Name: "LBU-74 Bomb", Tech: TechRequirements{Weapons: 15, Electronics: 12}, Mass: 45, Cost: Cost{14, 1, 33, 12}, KillRate: 4, StructureKill: 45},
	9:  {ID: 9, Name: "Hush-a-Boom", Tech: TechRequirements{Weapons: 12, Electronics: 12, Biotech: 12}, Mass: 5, Cost: Cost{5, 1, 5, 0}, KillRate: 30, StructureKill: 2},
	10: {ID: 10, Name: "Retro Bomb", Tech: TechRequirements{Weapons: 10, Biotech: 12}, Mass: 45, Cost: Cost{50, 15, 15, 10}, UnterraformRate: 1},
	11: {ID: 11, Name: "Smart Bomb", Tech: TechRequirements{Weapons: 5, Biotech: 7}, Mass: 50, Cost: Cost{27, 1, 22, 0}, KillRate: 13, IsSmart: true},
	12: {ID: 12, Name: "Neutron Bomb", Tech: TechRequirements{Weapons: 10, Biotech: 10}, Mass: 57, Cost: Cost{30, 1, 30, 0}, KillRate: 22, IsSmart: true},
	13: {ID: 13, Name: "Enriched Neutron Bomb", Tech: TechRequirements{Weapons: 15, Biotech: 12}, Mass: 64, Cost: Cost{25, 1, 36, 0}, KillRate: 35, IsSmart: true},
	14: {ID: 14, Name: "Peerless Bomb", Tech: TechRequirements{Weapons: 22, Biotech: 15}, Mass: 55, Cost: Cost{32, 1, 33, 0}, KillRate: 50, IsSmart: true},
	15: {ID: 15, Name: "Annihilator Bomb", Tech: TechRequirements{Weapons: 26, Biotech: 17}, Mass: 50, Cost: Cost{28, 1, 30, 0}, KillRate: 70, IsSmart: true},
}

// Orbital represents a stargate or mass driver
type Orbital struct {
	ID           int
	Name         string
	Tech         TechRequirements
	Mass         int
	Cost         Cost
	MassLimit    int // Max mass for stargate (-1 = any)
	RangeLimit   int // Max range for stargate (-1 = any)
	WarpSpeed    int // For mass drivers
	IsStargate   bool
	IsMassDriver bool
}

// Orbitals contains all orbital installation definitions
var Orbitals = map[int]*Orbital{
	1:  {ID: 1, Name: "Stargate 100/250", Tech: TechRequirements{Propulsion: 5, Construction: 5}, Mass: 0, Cost: Cost{400, 100, 40, 40}, MassLimit: 100, RangeLimit: 250, IsStargate: true},
	2:  {ID: 2, Name: "Stargate any/300", Tech: TechRequirements{Propulsion: 6, Construction: 10}, Mass: 0, Cost: Cost{500, 100, 40, 40}, MassLimit: -1, RangeLimit: 300, IsStargate: true},
	3:  {ID: 3, Name: "Stargate 150/600", Tech: TechRequirements{Propulsion: 11, Construction: 7}, Mass: 0, Cost: Cost{1000, 100, 40, 40}, MassLimit: 150, RangeLimit: 600, IsStargate: true},
	4:  {ID: 4, Name: "Stargate 300/500", Tech: TechRequirements{Propulsion: 9, Construction: 13}, Mass: 0, Cost: Cost{1200, 100, 40, 40}, MassLimit: 300, RangeLimit: 500, IsStargate: true},
	5:  {ID: 5, Name: "Stargate 100/any", Tech: TechRequirements{Propulsion: 16, Construction: 12}, Mass: 0, Cost: Cost{1400, 100, 40, 40}, MassLimit: 100, RangeLimit: -1, IsStargate: true},
	6:  {ID: 6, Name: "Stargate any/800", Tech: TechRequirements{Propulsion: 12, Construction: 18}, Mass: 0, Cost: Cost{1400, 100, 40, 40}, MassLimit: -1, RangeLimit: 800, IsStargate: true},
	7:  {ID: 7, Name: "Stargate any/any", Tech: TechRequirements{Propulsion: 19, Construction: 24}, Mass: 0, Cost: Cost{1600, 100, 40, 40}, MassLimit: -1, RangeLimit: -1, IsStargate: true},
	8:  {ID: 8, Name: "Mass Driver 5", Tech: TechRequirements{Energy: 4}, Mass: 48, Cost: Cost{140, 48, 40, 40}, WarpSpeed: 5, IsMassDriver: true},
	9:  {ID: 9, Name: "Mass Driver 6", Tech: TechRequirements{Energy: 7}, Mass: 48, Cost: Cost{288, 48, 40, 40}, WarpSpeed: 6, IsMassDriver: true},
	10: {ID: 10, Name: "Mass Driver 7", Tech: TechRequirements{Energy: 9}, Mass: 200, Cost: Cost{1024, 200, 200, 200}, WarpSpeed: 7, IsMassDriver: true},
	11: {ID: 11, Name: "Super Driver 8", Tech: TechRequirements{Energy: 11}, Mass: 48, Cost: Cost{512, 48, 40, 40}, WarpSpeed: 8, IsMassDriver: true},
	12: {ID: 12, Name: "Super Driver 9", Tech: TechRequirements{Energy: 13}, Mass: 48, Cost: Cost{648, 48, 40, 40}, WarpSpeed: 9, IsMassDriver: true},
	13: {ID: 13, Name: "Ultra Driver 10", Tech: TechRequirements{Energy: 15}, Mass: 200, Cost: Cost{1936, 200, 200, 200}, WarpSpeed: 10, IsMassDriver: true},
	14: {ID: 14, Name: "Ultra Driver 11", Tech: TechRequirements{Energy: 17}, Mass: 48, Cost: Cost{968, 48, 40, 40}, WarpSpeed: 11, IsMassDriver: true},
	15: {ID: 15, Name: "Ultra Driver 12", Tech: TechRequirements{Energy: 20}, Mass: 48, Cost: Cost{1152, 48, 40, 40}, WarpSpeed: 12, IsMassDriver: true},
	16: {ID: 16, Name: "Ultra Driver 13", Tech: TechRequirements{Energy: 24}, Mass: 48, Cost: Cost{1352, 48, 40, 40}, WarpSpeed: 13, IsMassDriver: true},
}

// Terraformer item IDs
const (
	TerraformTotal3      = 1
	TerraformTotal5      = 2
	TerraformTotal7      = 3
	TerraformTotal10     = 4
	TerraformTotal15     = 5
	TerraformTotal20     = 6
	TerraformTotal25     = 7
	TerraformTotal30     = 8
	TerraformGravity3    = 9
	TerraformGravity7    = 10
	TerraformGravity11   = 11
	TerraformGravity15   = 12
	TerraformTemp3       = 13
	TerraformTemp7       = 14
	TerraformTemp11      = 15
	TerraformTemp15      = 16
	TerraformRadiation3  = 17
	TerraformRadiation7  = 18
	TerraformRadiation11 = 19
	TerraformRadiation15 = 20
)

// Terraformer represents a terraforming technology
type Terraformer struct {
	ID            int
	Name          string
	Tech          TechRequirements
	Cost          Cost
	TerraformRate int    // Amount of terraforming per click
	TerraformType string // "Total", "Gravity", "Temp", or "Radiation"
}

// Terraformers contains all terraforming technology definitions
var Terraformers = map[int]*Terraformer{
	TerraformTotal3:      {ID: TerraformTotal3, Name: "Total Terraform +3", Tech: TechRequirements{}, Cost: Cost{0, 0, 0, 70}, TerraformRate: 3, TerraformType: "Total"},
	TerraformTotal5:      {ID: TerraformTotal5, Name: "Total Terraform +5", Tech: TechRequirements{Biotech: 3}, Cost: Cost{0, 0, 0, 70}, TerraformRate: 5, TerraformType: "Total"},
	TerraformTotal7:      {ID: TerraformTotal7, Name: "Total Terraform +7", Tech: TechRequirements{Biotech: 6}, Cost: Cost{0, 0, 0, 70}, TerraformRate: 7, TerraformType: "Total"},
	TerraformTotal10:     {ID: TerraformTotal10, Name: "Total Terraform +10", Tech: TechRequirements{Biotech: 9}, Cost: Cost{0, 0, 0, 70}, TerraformRate: 10, TerraformType: "Total"},
	TerraformTotal15:     {ID: TerraformTotal15, Name: "Total Terraform +15", Tech: TechRequirements{Biotech: 13}, Cost: Cost{0, 0, 0, 70}, TerraformRate: 15, TerraformType: "Total"},
	TerraformTotal20:     {ID: TerraformTotal20, Name: "Total Terraform +20", Tech: TechRequirements{Biotech: 17}, Cost: Cost{0, 0, 0, 70}, TerraformRate: 20, TerraformType: "Total"},
	TerraformTotal25:     {ID: TerraformTotal25, Name: "Total Terraform +25", Tech: TechRequirements{Biotech: 22}, Cost: Cost{0, 0, 0, 70}, TerraformRate: 25, TerraformType: "Total"},
	TerraformTotal30:     {ID: TerraformTotal30, Name: "Total Terraform +30", Tech: TechRequirements{Biotech: 25}, Cost: Cost{0, 0, 0, 70}, TerraformRate: 30, TerraformType: "Total"},
	TerraformGravity3:    {ID: TerraformGravity3, Name: "Gravity Terraform +3", Tech: TechRequirements{Propulsion: 1, Biotech: 1}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 3, TerraformType: "Gravity"},
	TerraformGravity7:    {ID: TerraformGravity7, Name: "Gravity Terraform +7", Tech: TechRequirements{Propulsion: 5, Biotech: 2}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 7, TerraformType: "Gravity"},
	TerraformGravity11:   {ID: TerraformGravity11, Name: "Gravity Terraform +11", Tech: TechRequirements{Propulsion: 10, Biotech: 3}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 11, TerraformType: "Gravity"},
	TerraformGravity15:   {ID: TerraformGravity15, Name: "Gravity Terraform +15", Tech: TechRequirements{Propulsion: 16, Biotech: 4}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 15, TerraformType: "Gravity"},
	TerraformTemp3:       {ID: TerraformTemp3, Name: "Temp Terraform +3", Tech: TechRequirements{Energy: 1, Biotech: 1}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 3, TerraformType: "Temp"},
	TerraformTemp7:       {ID: TerraformTemp7, Name: "Temp Terraform +7", Tech: TechRequirements{Energy: 5, Biotech: 2}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 7, TerraformType: "Temp"},
	TerraformTemp11:      {ID: TerraformTemp11, Name: "Temp Terraform +11", Tech: TechRequirements{Energy: 10, Biotech: 3}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 11, TerraformType: "Temp"},
	TerraformTemp15:      {ID: TerraformTemp15, Name: "Temp Terraform +15", Tech: TechRequirements{Energy: 16, Biotech: 4}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 15, TerraformType: "Temp"},
	TerraformRadiation3:  {ID: TerraformRadiation3, Name: "Radiation Terraform +3", Tech: TechRequirements{Weapons: 1, Biotech: 1}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 3, TerraformType: "Radiation"},
	TerraformRadiation7:  {ID: TerraformRadiation7, Name: "Radiation Terraform +7", Tech: TechRequirements{Weapons: 5, Biotech: 2}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 7, TerraformType: "Radiation"},
	TerraformRadiation11: {ID: TerraformRadiation11, Name: "Radiation Terraform +11", Tech: TechRequirements{Weapons: 10, Biotech: 3}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 11, TerraformType: "Radiation"},
	TerraformRadiation15: {ID: TerraformRadiation15, Name: "Radiation Terraform +15", Tech: TechRequirements{Weapons: 16, Biotech: 4}, Cost: Cost{0, 0, 0, 100}, TerraformRate: 15, TerraformType: "Radiation"},
}

// GetTerraformer returns a terraformer by ID
func GetTerraformer(id int) *Terraformer { return Terraformers[id] }

// Planetary defense item IDs
const (
	DefenseSDI             = 10
	DefenseMissileBattery  = 11
	DefenseLaserBattery    = 12
	DefensePlanetaryShield = 13
	DefenseNeutronShield   = 14
	DefenseGenesisDevice   = 15
)

// PlanetaryDefense represents a planetary defense installation
type PlanetaryDefense struct {
	ID              int
	Name            string
	Tech            TechRequirements
	Cost            Cost
	DefenseValue    int  // Defense coverage percentage
	IsGenesisDevice bool // Special - creates planets
}

// PlanetaryDefenses contains all planetary defense definitions
var PlanetaryDefenses = map[int]*PlanetaryDefense{
	DefenseSDI:             {ID: DefenseSDI, Name: "SDI", Tech: TechRequirements{}, Cost: Cost{15, 5, 5, 5}, DefenseValue: 10},
	DefenseMissileBattery:  {ID: DefenseMissileBattery, Name: "Missile Battery", Tech: TechRequirements{Energy: 5}, Cost: Cost{15, 5, 5, 5}, DefenseValue: 20},
	DefenseLaserBattery:    {ID: DefenseLaserBattery, Name: "Laser Battery", Tech: TechRequirements{Energy: 10}, Cost: Cost{15, 5, 5, 5}, DefenseValue: 24},
	DefensePlanetaryShield: {ID: DefensePlanetaryShield, Name: "Planetary Shield", Tech: TechRequirements{Energy: 16}, Cost: Cost{15, 5, 5, 5}, DefenseValue: 30},
	DefenseNeutronShield:   {ID: DefenseNeutronShield, Name: "Neutron Shield", Tech: TechRequirements{Energy: 23}, Cost: Cost{15, 5, 5, 5}, DefenseValue: 38},
	DefenseGenesisDevice:   {ID: DefenseGenesisDevice, Name: "Genesis Device", Tech: TechRequirements{Energy: 20, Weapons: 10, Propulsion: 10, Construction: 20, Electronics: 10, Biotech: 20}, Cost: Cost{5000, 0, 0, 0}, IsGenesisDevice: true},
}

// GetPlanetaryDefense returns a planetary defense by ID
func GetPlanetaryDefense(id int) *PlanetaryDefense { return PlanetaryDefenses[id] }

// Lookup functions

// GetEngine returns an engine by ID
func GetEngine(id int) *Engine { return Engines[id] }

// GetShield returns a shield by ID
func GetShield(id int) *Shield { return Shields[id] }

// GetArmor returns armor by ID
func GetArmor(id int) *Armor { return Armors[id] }

// GetBeamWeapon returns a beam weapon by ID
func GetBeamWeapon(id int) *BeamWeapon { return BeamWeapons[id] }

// GetTorpedo returns a torpedo by ID
func GetTorpedo(id int) *Torpedo { return Torpedoes[id] }

// GetElectrical returns an electrical component by ID
func GetElectrical(id int) *Electrical { return Electricals[id] }

// GetMechanical returns a mechanical component by ID
func GetMechanical(id int) *Mechanical { return Mechanicals[id] }

// GetMineLayer returns a mine layer by ID
func GetMineLayer(id int) *MineLayer { return MineLayers[id] }

// GetMiningRobot returns a mining robot by ID
func GetMiningRobot(id int) *MiningRobot { return MiningRobots[id] }

// GetBomb returns a bomb by ID
func GetBomb(id int) *Bomb { return Bombs[id] }

// GetOrbital returns an orbital installation by ID
func GetOrbital(id int) *Orbital { return Orbitals[id] }
