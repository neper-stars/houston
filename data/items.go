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

// ScannerStats holds the range capabilities and requirements of a scanner.
type ScannerStats struct {
	NormalRange      int  // Normal scanning range in light-years
	PenetratingRange int  // Penetrating scanning range in light-years
	StealsCargo      bool // Can detect enemy cargo (Pick Pocket, Robber Baron)
	ElectronicsLevel int  // Electronics tech level required
}

// ShipScannerStats maps ship scanner item IDs to their stats.
var ShipScannerStats = map[int]ScannerStats{
	ScannerBat:         {NormalRange: 0, PenetratingRange: 0},
	ScannerRhino:       {NormalRange: 50, PenetratingRange: 0},
	ScannerMole:        {NormalRange: 100, PenetratingRange: 0},
	ScannerDNA:         {NormalRange: 125, PenetratingRange: 0},
	ScannerPossum:      {NormalRange: 150, PenetratingRange: 0},
	ScannerPickPocket:  {NormalRange: 80, PenetratingRange: 0, StealsCargo: true},
	ScannerChameleon:   {NormalRange: 160, PenetratingRange: 45},
	ScannerFerret:      {NormalRange: 185, PenetratingRange: 50},
	ScannerDolphin:     {NormalRange: 220, PenetratingRange: 100},
	ScannerGazelle:     {NormalRange: 225, PenetratingRange: 0},
	ScannerRNA:         {NormalRange: 230, PenetratingRange: 0},
	ScannerCheetah:     {NormalRange: 275, PenetratingRange: 0},
	ScannerElephant:    {NormalRange: 300, PenetratingRange: 200},
	ScannerEagleEye:    {NormalRange: 335, PenetratingRange: 0},
	ScannerRobberBaron: {NormalRange: 220, PenetratingRange: 120, StealsCargo: true},
	ScannerPeerless:    {NormalRange: 500, PenetratingRange: 0},
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

// PlanetaryScannerStats maps planetary scanner item IDs to their stats.
// ElectronicsLevel is the tech required to build this scanner.
// For Snooper scanners, the name refers to the normal range; penetrating is half that.
var PlanetaryScannerStats = map[int]ScannerStats{
	PlanetaryScannerViewer50:   {NormalRange: 50, PenetratingRange: 0, ElectronicsLevel: 0},
	PlanetaryScannerViewer90:   {NormalRange: 90, PenetratingRange: 0, ElectronicsLevel: 1},
	PlanetaryScannerScoper150:  {NormalRange: 150, PenetratingRange: 0, ElectronicsLevel: 3},
	PlanetaryScannerScoper220:  {NormalRange: 220, PenetratingRange: 0, ElectronicsLevel: 6},
	PlanetaryScannerScoper280:  {NormalRange: 280, PenetratingRange: 0, ElectronicsLevel: 8},
	PlanetaryScannerSnooper320: {NormalRange: 320, PenetratingRange: 160, ElectronicsLevel: 10},
	PlanetaryScannerSnooper400: {NormalRange: 400, PenetratingRange: 200, ElectronicsLevel: 13},
	PlanetaryScannerSnooper500: {NormalRange: 500, PenetratingRange: 250, ElectronicsLevel: 16},
	PlanetaryScannerSnooper620: {NormalRange: 620, PenetratingRange: 310, ElectronicsLevel: 20},
}

// GetBestPlanetaryScanner returns the best planetary scanner available at the given electronics tech level.
// Returns the scanner stats and the scanner ID.
func GetBestPlanetaryScanner(electronicsLevel int) (ScannerStats, int) {
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
		stats := PlanetaryScannerStats[id]
		if electronicsLevel >= stats.ElectronicsLevel {
			return stats, id
		}
	}

	// Fallback to Viewer 50 (always available)
	return PlanetaryScannerStats[PlanetaryScannerViewer50], PlanetaryScannerViewer50
}

// JoATIntrinsicScanner returns the intrinsic scanner range for Jack of All Trades ships.
// JoAT ships have built-in scanners that improve with Electronics tech level.
// Formula: Normal range = Electronics × 20, Penetrating range = Electronics × 10
// Minimum ranges are 60 ly normal and 30 ly penetrating (equivalent to Electronics 3).
func JoATIntrinsicScanner(electronicsLevel int) ScannerStats {
	normalRange := electronicsLevel * 20
	penRange := electronicsLevel * 10

	// Apply minimums (60/30 ly, equivalent to Electronics level 3)
	if normalRange < 60 {
		normalRange = 60
	}
	if penRange < 30 {
		penRange = 30
	}

	return ScannerStats{
		NormalRange:      normalRange,
		PenetratingRange: penRange,
	}
}

// PlanetaryScannerNames maps planetary scanner IDs to display names.
var PlanetaryScannerNames = map[int]string{
	PlanetaryScannerViewer50:   "Viewer 50",
	PlanetaryScannerViewer90:   "Viewer 90",
	PlanetaryScannerScoper150:  "Scoper 150",
	PlanetaryScannerScoper220:  "Scoper 220",
	PlanetaryScannerScoper280:  "Scoper 280",
	PlanetaryScannerSnooper320: "Snooper 320X",
	PlanetaryScannerSnooper400: "Snooper 400X",
	PlanetaryScannerSnooper500: "Snooper 500X",
	PlanetaryScannerSnooper620: "Snooper 620X",
}

// GetShipScannerStats returns the scanner stats for a ship scanner ID.
func GetShipScannerStats(scannerID int) (ScannerStats, bool) {
	stats, ok := ShipScannerStats[scannerID]
	return stats, ok
}

// GetPlanetaryScannerStats returns the scanner stats for a planetary scanner ID.
func GetPlanetaryScannerStats(scannerID int) (ScannerStats, bool) {
	stats, ok := PlanetaryScannerStats[scannerID]
	return stats, ok
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
const (
	MechColonizationModule        = 1
	MechOrbitalConstructionModule = 2
	MechCargoPod                  = 3
	MechSuperCargoPod             = 4
	MechMultiCargoPod             = 5
	MechFuelTank                  = 6
	MechSuperFuelTank             = 7
	MechManeuveringJet            = 8
	MechOverthruster              = 9
	MechJumpGate                  = 10
	MechBeamDeflector             = 11
)

// Mining robot item IDs
const (
	MiningRoboMidget = 1
	MiningRoboMini   = 2
	MiningRobo       = 3
	MiningRoboMaxi   = 4
	MiningRoboSuper  = 5
	MiningRoboUltra  = 6
	MiningAlien      = 7
	MiningOrbitalAdj = 8
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
