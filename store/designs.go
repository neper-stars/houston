package store

import (
	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/data"
)

// DesignEntity represents a ship or starbase design.
type DesignEntity struct {
	meta EntityMeta

	// Identification
	DesignNumber int  // Slot 0-15
	Owner        int  // Player index
	IsStarbase   bool // True if starbase design

	// Design info
	Name   string
	HullId int // Hull type ID (see data.Hull* constants)

	// Raw block (preserved for re-encoding)
	designBlock *blocks.DesignBlock
}

// Meta returns the entity metadata.
func (d *DesignEntity) Meta() *EntityMeta {
	return &d.meta
}

// RawBlocks returns the original blocks.
func (d *DesignEntity) RawBlocks() []blocks.Block {
	if d.designBlock != nil {
		return []blocks.Block{*d.designBlock}
	}
	return nil
}

// SetDirty marks the entity as modified.
func (d *DesignEntity) SetDirty() {
	d.meta.Dirty = true
}

// newDesignEntityFromBlock creates a DesignEntity from a DesignBlock.
// The owner is taken from the source file's player index.
func newDesignEntityFromBlock(db *blocks.DesignBlock, source *FileSource) *DesignEntity {
	entityType := EntityTypeDesign
	if db.IsStarbase {
		entityType = EntityTypeStarbaseDesign
	}

	owner := source.PlayerIndex

	// Full designs (with component info) have higher quality than partial designs
	quality := QualityFull
	if !db.IsFullDesign {
		quality = QualityPartial
	}

	entity := &DesignEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   entityType,
				Owner:  owner,
				Number: db.DesignNumber,
			},
			BestSource: source,
			Quality:    quality,
			Turn:       source.Turn,
		},
		DesignNumber: db.DesignNumber,
		Owner:        owner,
		IsStarbase:   db.IsStarbase,
		Name:         db.Name,
		HullId:       db.HullId,
		designBlock:  db,
	}
	entity.meta.AddSource(source)
	return entity
}

// GetScannerRanges returns the best normal and penetrating scanner ranges
// from scanners equipped on this design.
//
// The Category field indicates the item type equipped (ItemCategoryScanner = 0x0002).
// ItemId is 0-indexed, so we add 1 to get the scanner constant (ScannerBat=1, etc.).
//
// Returns (0, 0) if no scanners are equipped.
func (d *DesignEntity) GetScannerRanges() (normal, penetrating int) {
	if d.designBlock == nil {
		return 0, 0
	}

	bestNormal := 0
	bestPen := 0

	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryScanner {
			continue
		}

		// ItemId is 0-indexed, scanner constants are 1-indexed
		scannerID := slot.ItemId + 1
		scanner := data.GetScanner(scannerID)

		if scanner != nil {
			if scanner.NormalRange > bestNormal {
				bestNormal = scanner.NormalRange
			}
			if scanner.PenetratingRange > bestPen {
				bestPen = scanner.PenetratingRange
			}
		}
	}

	return bestNormal, bestPen
}

// GetNormalScannerRange returns the best normal scanner range from equipped scanners.
func (d *DesignEntity) GetNormalScannerRange() int {
	normal, _ := d.GetScannerRanges()
	return normal
}

// GetPenetratingScannerRange returns the best penetrating scanner range from equipped scanners.
func (d *DesignEntity) GetPenetratingScannerRange() int {
	_, pen := d.GetScannerRanges()
	return pen
}

// HasScanner returns true if this design has any scanner equipped.
func (d *DesignEntity) HasScanner() bool {
	normal, pen := d.GetScannerRanges()
	return normal > 0 || pen > 0
}

// Hull returns the hull definition for this design.
// Returns nil if the hull ID is unknown.
func (d *DesignEntity) Hull() *data.Hull {
	return data.Hulls[d.HullId]
}

// EquippedItem represents an item equipped in a design slot.
type EquippedItem struct {
	SlotIndex int    // Index in the design's slot array
	Category  uint16 // Item category (blocks.ItemCategory*)
	ItemID    int    // 1-indexed item ID (for data lookup)
	Count     int    // Number of items in this slot
}

// EquippedItems returns all non-empty slots in this design.
func (d *DesignEntity) EquippedItems() []EquippedItem {
	if d.designBlock == nil {
		return nil
	}

	var items []EquippedItem
	for i, slot := range d.designBlock.Slots {
		if slot.Count == 0 {
			continue
		}
		items = append(items, EquippedItem{
			SlotIndex: i,
			Category:  slot.Category,
			ItemID:    slot.ItemId + 1, // Convert to 1-indexed
			Count:     int(slot.Count),
		})
	}
	return items
}

// ItemsByCategory returns all equipped items of a specific category.
func (d *DesignEntity) ItemsByCategory(category uint16) []EquippedItem {
	if d.designBlock == nil {
		return nil
	}

	var items []EquippedItem
	for i, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != category {
			continue
		}
		items = append(items, EquippedItem{
			SlotIndex: i,
			Category:  slot.Category,
			ItemID:    slot.ItemId + 1,
			Count:     int(slot.Count),
		})
	}
	return items
}

// GetCloakPercent returns the total cloaking units for this design.
// Cloaking sources include:
// - Electrical: Transport Cloaking (300, freighters only), Stealth (70), Super-Stealth (140), Ultra-Stealth (540)
// - Shields: Shadow Shield (70)
// - Armor: Depleted Neutronium (50)
// - Scanners: Chameleon Scanner (40)
// The returned value is "cloak units" which get weighted by ship mass in fleet calculations.
func (d *DesignEntity) GetCloakPercent() int {
	if d.designBlock == nil {
		return 0
	}

	totalCloak := 0
	isFreighter := d.isFreighterHull()

	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 {
			continue
		}

		itemID := slot.ItemId + 1
		count := int(slot.Count)

		switch slot.Category {
		case blocks.ItemCategoryElectrical:
			elec := data.GetElectrical(itemID)
			if elec == nil || elec.CloakPercent == 0 {
				continue
			}
			// Transport Cloaking only works on freighter hulls
			if itemID == data.ElecTransportCloaking && !isFreighter {
				continue
			}
			totalCloak += elec.CloakPercent * count

		case blocks.ItemCategoryShield:
			shield := data.GetShield(itemID)
			if shield != nil && shield.CloakPercent > 0 {
				totalCloak += shield.CloakPercent * count
			}

		case blocks.ItemCategoryArmor:
			armor := data.GetArmor(itemID)
			if armor != nil && armor.CloakPercent > 0 {
				totalCloak += armor.CloakPercent * count
			}

		case blocks.ItemCategoryScanner:
			scanner := data.GetScanner(itemID)
			if scanner != nil && scanner.CloakPercent > 0 {
				totalCloak += scanner.CloakPercent * count
			}
		}
	}

	return totalCloak
}

// isFreighterHull returns true if this design uses a freighter hull.
func (d *DesignEntity) isFreighterHull() bool {
	switch d.HullId {
	case data.HullSmallFreighter, data.HullMediumFreighter,
		data.HullLargeFreighter, data.HullSuperFreighter:
		return true
	}
	return false
}

// HasCloak returns true if this design has any cloaking device.
func (d *DesignEntity) HasCloak() bool {
	return d.GetCloakPercent() > 0
}

// GetTachyonCount returns the number of Tachyon Detectors equipped on this design.
// Tachyon Detectors reduce the effective cloaking of enemy fleets.
func (d *DesignEntity) GetTachyonCount() int {
	if d.designBlock == nil {
		return 0
	}

	for _, slot := range d.designBlock.Slots {
		if slot.Count > 0 && slot.Category == blocks.ItemCategoryElectrical {
			if slot.ItemId+1 == data.ElecTachyonDetector {
				return int(slot.Count)
			}
		}
	}
	return 0
}

// HasTachyonDetector returns true if this design has any Tachyon Detectors.
func (d *DesignEntity) HasTachyonDetector() bool {
	return d.GetTachyonCount() > 0
}

// GetMinesweepRate returns the total minesweeping rate for this design.
// Minesweeping is done by beam weapons. Each beam weapon sweeps mines
// at a rate of: (weapon power) * (count) * (gattling bonus if applicable)
// Gattling weapons (Mini Gun, Gatling Gun, etc.) are 4x more effective.
func (d *DesignEntity) GetMinesweepRate() int {
	if d.designBlock == nil {
		return 0
	}

	totalSweep := 0

	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryBeamWeapon {
			continue
		}

		itemID := slot.ItemId + 1
		beam := data.GetBeamWeapon(itemID)
		if beam == nil {
			continue
		}

		// Minesweeping rate = power * count
		// Gattling weapons are 4x effective at sweeping
		rate := beam.Power * int(slot.Count)
		if beam.IsGatling {
			rate *= 4
		}
		totalSweep += rate
	}

	return totalSweep
}

// HasMinesweep returns true if this design can sweep mines.
func (d *DesignEntity) HasMinesweep() bool {
	return d.GetMinesweepRate() > 0
}

// GetEngine returns the engine equipped on this design.
// Returns nil if no engine is equipped (shouldn't happen for valid designs).
func (d *DesignEntity) GetEngine() *data.Engine {
	if d.designBlock == nil {
		return nil
	}

	for _, slot := range d.designBlock.Slots {
		if slot.Count > 0 && slot.Category == blocks.ItemCategoryEngine {
			return data.GetEngine(slot.ItemId + 1)
		}
	}
	return nil
}

// GetTotalShieldValue returns the total shield strength for this design.
// This is the sum of (shield value * count) for all equipped shields.
func (d *DesignEntity) GetTotalShieldValue() int {
	if d.designBlock == nil {
		return 0
	}

	total := 0
	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryShield {
			continue
		}

		shield := data.GetShield(slot.ItemId + 1)
		if shield != nil {
			total += shield.ShieldValue * int(slot.Count)
		}
	}
	return total
}

// GetTotalArmorValue returns the total armor for this design.
// This includes hull armor plus equipped armor items.
func (d *DesignEntity) GetTotalArmorValue() int {
	total := 0

	// Base hull armor
	if hull := d.Hull(); hull != nil {
		total += hull.Armor
	}

	if d.designBlock == nil {
		return total
	}

	// Add equipped armor
	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryArmor {
			continue
		}

		armor := data.GetArmor(slot.ItemId + 1)
		if armor != nil {
			total += armor.ArmorValue * int(slot.Count)
		}
	}

	// Some shields also provide armor (Croby Sharmor, Langston Shell)
	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryShield {
			continue
		}

		shield := data.GetShield(slot.ItemId + 1)
		if shield != nil && shield.ArmorValue > 0 {
			total += shield.ArmorValue * int(slot.Count)
		}
	}

	// Multi Cargo Pod also provides armor
	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryMechanical {
			continue
		}

		mech := data.GetMechanical(slot.ItemId + 1)
		if mech != nil && mech.ArmorValue > 0 {
			total += mech.ArmorValue * int(slot.Count)
		}
	}

	return total
}

// GetCargoCapacity returns the total cargo capacity for this design.
// This includes hull cargo capacity plus cargo pods.
func (d *DesignEntity) GetCargoCapacity() int {
	total := 0

	// Base hull cargo
	if hull := d.Hull(); hull != nil {
		total += hull.CargoCapacity
	}

	if d.designBlock == nil {
		return total
	}

	// Add cargo pods
	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryMechanical {
			continue
		}

		mech := data.GetMechanical(slot.ItemId + 1)
		if mech != nil && mech.CargoCapacity > 0 {
			total += mech.CargoCapacity * int(slot.Count)
		}
	}

	return total
}

// GetFuelCapacity returns the total fuel capacity for this design.
// This includes hull fuel capacity plus fuel tanks.
func (d *DesignEntity) GetFuelCapacity() int {
	total := 0

	// Base hull fuel
	if hull := d.Hull(); hull != nil {
		total += hull.FuelCapacity
	}

	if d.designBlock == nil {
		return total
	}

	// Add fuel tanks
	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryMechanical {
			continue
		}

		mech := data.GetMechanical(slot.ItemId + 1)
		if mech != nil && mech.FuelCapacity > 0 {
			total += mech.FuelCapacity * int(slot.Count)
		}
	}

	return total
}

// GetMiningRate returns the total mining rate for this design.
// Only applies to ships with mining robots equipped.
func (d *DesignEntity) GetMiningRate() int {
	if d.designBlock == nil {
		return 0
	}

	total := 0
	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryMiningRobot {
			continue
		}

		robot := data.GetMiningRobot(slot.ItemId + 1)
		if robot != nil {
			total += robot.MiningRate * int(slot.Count)
		}
	}

	return total
}

// HasMining returns true if this design has mining capability.
func (d *DesignEntity) HasMining() bool {
	return d.GetMiningRate() > 0
}

// GetMinelayingRate returns the mines per year this design can lay, by mine type.
// Returns (normal, heavy, speed) mine laying rates.
func (d *DesignEntity) GetMinelayingRate() (normal, heavy, speed int) {
	if d.designBlock == nil {
		return 0, 0, 0
	}

	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryMineLayer {
			continue
		}

		layer := data.GetMineLayer(slot.ItemId + 1)
		if layer == nil {
			continue
		}

		rate := layer.MinesPerYear * int(slot.Count)
		switch layer.MineType {
		case "Normal":
			normal += rate
		case "Heavy":
			heavy += rate
		case "Speed":
			speed += rate
		}
	}

	return normal, heavy, speed
}

// HasMinelaying returns true if this design can lay mines.
func (d *DesignEntity) HasMinelaying() bool {
	n, h, s := d.GetMinelayingRate()
	return n > 0 || h > 0 || s > 0
}

// CanColonize returns true if this design has a colonization module.
func (d *DesignEntity) CanColonize() bool {
	if d.designBlock == nil {
		return false
	}

	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryMechanical {
			continue
		}

		mech := data.GetMechanical(slot.ItemId + 1)
		if mech != nil && mech.Colonizer {
			return true
		}
	}
	return false
}

// CanBuildStarbase returns true if this design has an orbital construction module.
func (d *DesignEntity) CanBuildStarbase() bool {
	if d.designBlock == nil {
		return false
	}

	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryMechanical {
			continue
		}

		mech := data.GetMechanical(slot.ItemId + 1)
		if mech != nil && mech.OrbitalBuild {
			return true
		}
	}
	return false
}

// GetBombCapability returns bomb stats for this design.
// Returns total kill rate (min colonists killed per bomb run) and structure kill rate.
func (d *DesignEntity) GetBombCapability() (killRate, structureKill int, hasSmart bool) {
	if d.designBlock == nil {
		return 0, 0, false
	}

	for _, slot := range d.designBlock.Slots {
		if slot.Count == 0 || slot.Category != blocks.ItemCategoryBomb {
			continue
		}

		bomb := data.GetBomb(slot.ItemId + 1)
		if bomb == nil {
			continue
		}

		killRate += bomb.KillRate * int(slot.Count)
		structureKill += bomb.StructureKill * int(slot.Count)
		if bomb.IsSmart {
			hasSmart = true
		}
	}

	return killRate, structureKill, hasSmart
}

// HasBombs returns true if this design has bombing capability.
func (d *DesignEntity) HasBombs() bool {
	k, _, _ := d.GetBombCapability()
	return k > 0
}

// DesignMap is a convenience type for looking up designs by slot.
type DesignMap map[int]*DesignEntity
