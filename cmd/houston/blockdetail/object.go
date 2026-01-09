package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.ObjectBlockType, FormatObject)
}

// objectTypeName returns human-readable name for object type
func objectTypeName(objType int) string {
	names := map[int]string{
		blocks.ObjectTypeMinefield:     "Minefield",
		blocks.ObjectTypePacketSalvage: "Packet/Salvage",
		blocks.ObjectTypeWormhole:      "Wormhole",
		blocks.ObjectTypeMysteryTrader: "Mystery Trader",
	}
	if name, ok := names[objType]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", objType)
}

// minefieldTypeName returns human-readable name for minefield type
func minefieldTypeName(mfType int) string {
	names := map[int]string{
		blocks.MinefieldTypeStandard:  "Standard",
		blocks.MinefieldTypeHeavy:     "Heavy",
		blocks.MinefieldTypeSpeedBump: "Speed Bump",
	}
	if name, ok := names[mfType]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", mfType)
}

// traderItemName returns human-readable name for trader item bit
func traderItemName(itemBit uint16) string {
	names := map[uint16]string{
		blocks.TraderItemMultiCargoPod:          "Multi Cargo Pod",
		blocks.TraderItemMultiFunctionPod:       "Multi Function Pod",
		blocks.TraderItemLangstonShield:         "Langston Shield",
		blocks.TraderItemMegaPolyShell:          "Mega Poly Shell",
		blocks.TraderItemAlienMiner:             "Alien Miner",
		blocks.TraderItemHushABoom:              "Hush-a-Boom",
		blocks.TraderItemAntiMatterTorpedo:      "Anti-Matter Torpedo",
		blocks.TraderItemMultiContainedMunition: "Multi Contained Munition",
		blocks.TraderItemMiniMorph:              "Mini Morph",
		blocks.TraderItemEnigmaPulsar:           "Enigma Pulsar",
		blocks.TraderItemGenesisDevice:          "Genesis Device",
		blocks.TraderItemJumpGate:               "Jump Gate",
		blocks.TraderItemShip:                   "Ship",
	}
	if name, ok := names[itemBit]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(0x%04X)", itemBit)
}

// FormatObject provides detailed view for ObjectBlock (type 43)
func FormatObject(block blocks.Block, index int) string {
	width := DefaultWidth
	ob, ok := block.(blocks.ObjectBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := ob.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	// Count object (2 bytes)
	if ob.IsCountObject {
		fields = append(fields, "── Count Object ──")
		fields = append(fields, FormatFieldRaw(0x00, 0x01, "Count",
			fmt.Sprintf("0x%02X%02X", d[1], d[0]),
			fmt.Sprintf("uint16 LE = %d objects", ob.Count)))

		fields = append(fields, "")
		fields = append(fields, "── Summary ──")
		fields = append(fields, fmt.Sprintf("  Object count: %d", ob.Count))

		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	if len(d) < 6 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Common fields
	objectWord := encoding.Read16(d, 0)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "ObjectId",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = 0x%04X", objectWord)))
	fields = append(fields, fmt.Sprintf("           %s bits 0-8: Number = %d",
		TreeBranch, ob.Number))
	fields = append(fields, fmt.Sprintf("           %s bits 9-12: Owner = Player %d",
		TreeBranch, ob.Owner+1))
	fields = append(fields, fmt.Sprintf("           %s bits 13-15: Type = %d (%s)",
		TreeEnd, ob.ObjectType, objectTypeName(ob.ObjectType)))

	fields = append(fields, FormatFieldRaw(0x02, 0x03, "X",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = %d", ob.X)))

	fields = append(fields, FormatFieldRaw(0x04, 0x05, "Y",
		fmt.Sprintf("0x%02X%02X", d[5], d[4]),
		fmt.Sprintf("uint16 LE = %d", ob.Y)))

	// Type-specific fields
	switch ob.ObjectType {
	case blocks.ObjectTypeMinefield:
		formatMinefield(&fields, ob, d)
	case blocks.ObjectTypePacketSalvage:
		formatPacket(&fields, ob, d)
	case blocks.ObjectTypeWormhole:
		formatWormhole(&fields, ob, d)
	case blocks.ObjectTypeMysteryTrader:
		formatMysteryTrader(&fields, ob, d)
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

func formatMinefield(fields *[]string, ob blocks.ObjectBlock, d []byte) {
	if len(d) < 14 {
		return
	}

	*fields = append(*fields, "")
	*fields = append(*fields, "── Minefield Data ──")

	// Bytes 6-9: Mine count (uint32)
	*fields = append(*fields, FormatFieldRaw(0x06, 0x09, "MineCount",
		fmt.Sprintf("0x%02X%02X%02X%02X", d[9], d[8], d[7], d[6]),
		fmt.Sprintf("uint32 LE = %d mines", ob.MineCount)))

	// Bytes 10-11: Visibility mask
	*fields = append(*fields, FormatFieldRaw(0x0A, 0x0B, "CanSeeBits",
		fmt.Sprintf("0x%02X%02X", d[11], d[10]),
		fmt.Sprintf("bitmask = 0x%04X", ob.MineCanSeeBits)))

	// Show which players can see
	var viewers []string
	for i := 0; i < 16; i++ {
		if ob.PlayerCanSeeMinefield(i) {
			viewers = append(viewers, fmt.Sprintf("P%d", i+1))
		}
	}
	if len(viewers) > 0 {
		*fields = append(*fields, fmt.Sprintf("           %s Visible to: %v", TreeEnd, viewers))
	}

	// Byte 12: Minefield type
	*fields = append(*fields, FormatFieldRaw(0x0C, 0x0C, "MinefieldType",
		fmt.Sprintf("0x%02X", d[12]),
		fmt.Sprintf("%d = %s", ob.MinefieldType, minefieldTypeName(ob.MinefieldType))))

	// Byte 13: Detonating flag
	detonating := "false"
	if ob.Detonating {
		detonating = "true"
	}
	*fields = append(*fields, FormatFieldRaw(0x0D, 0x0D, "Detonating",
		fmt.Sprintf("0x%02X", d[13]),
		detonating))

	// Bytes 14-15: Current visibility bits
	if len(d) >= 16 {
		*fields = append(*fields, FormatFieldRaw(0x0E, 0x0F, "CurrentSeeBits",
			fmt.Sprintf("0x%02X%02X", d[15], d[14]),
			fmt.Sprintf("bitmask = 0x%04X (grbitPlrNow)", ob.MineCurrentSeeBits)))

		// Show which players can currently see
		var currentViewers []string
		for i := 0; i < 16; i++ {
			if (ob.MineCurrentSeeBits & (1 << i)) != 0 {
				currentViewers = append(currentViewers, fmt.Sprintf("P%d", i+1))
			}
		}
		if len(currentViewers) > 0 {
			*fields = append(*fields, fmt.Sprintf("           %s Currently visible to: %v", TreeEnd, currentViewers))
		}
	}

	// Bytes 16-17: Turn number
	if len(d) >= 18 {
		*fields = append(*fields, FormatFieldRaw(0x10, 0x11, "TurnNumber",
			fmt.Sprintf("0x%02X%02X", d[17], d[16]),
			fmt.Sprintf("uint16 LE = %d", ob.MineTurnNumber)))
	}

	// Summary
	*fields = append(*fields, "")
	*fields = append(*fields, "── Summary ──")
	*fields = append(*fields, fmt.Sprintf("  %s minefield #%d @ (%d, %d)",
		minefieldTypeName(ob.MinefieldType), ob.Number, ob.X, ob.Y))
	*fields = append(*fields, fmt.Sprintf("  Owner: Player %d", ob.Owner+1))
	*fields = append(*fields, fmt.Sprintf("  Mines: %d", ob.MineCount))
	if ob.Detonating {
		*fields = append(*fields, "  Status: DETONATING")
	}
}

func formatPacket(fields *[]string, ob blocks.ObjectBlock, d []byte) {
	if len(d) < 14 {
		return
	}

	*fields = append(*fields, "")

	if ob.IsSalvageObject {
		*fields = append(*fields, "── Salvage Data ──")

		// Byte 6: 0xFF marker
		*fields = append(*fields, FormatFieldRaw(0x06, 0x06, "Marker",
			fmt.Sprintf("0x%02X", d[6]),
			"0xFF = salvage marker"))

		// Byte 7: Source fleet info
		*fields = append(*fields, FormatFieldRaw(0x07, 0x07, "SourceFleet",
			fmt.Sprintf("0x%02X", d[7]),
			fmt.Sprintf("byte value = 0x%02X", d[7])))
		*fields = append(*fields, fmt.Sprintf("           %s low nibble = Fleet #%d (0-indexed: %d)",
			TreeBranch, ob.SourceFleetID+1, ob.SourceFleetID))
		*fields = append(*fields, fmt.Sprintf("           %s high nibble = flags: 0x%X",
			TreeEnd, ob.SalvageSourceFlags))
	} else {
		*fields = append(*fields, "── Mineral Packet Data ──")

		// Byte 6: Destination planet
		*fields = append(*fields, FormatFieldRaw(0x06, 0x06, "DestPlanet",
			fmt.Sprintf("0x%02X", d[6]),
			fmt.Sprintf("Planet #%d", ob.DestinationPlanetID+1)))

		// Byte 7: Speed byte
		warpSpeed := ob.WarpSpeed()
		*fields = append(*fields, FormatFieldRaw(0x07, 0x07, "Speed",
			fmt.Sprintf("0x%02X", d[7]),
			fmt.Sprintf("(byte >> 2) - 44 = Warp %d", warpSpeed)))
	}

	// Bytes 8-9: Ironium
	*fields = append(*fields, FormatFieldRaw(0x08, 0x09, "Ironium",
		fmt.Sprintf("0x%02X%02X", d[9], d[8]),
		fmt.Sprintf("uint16 LE = %d kT", ob.Ironium)))

	// Bytes 10-11: Boranium
	*fields = append(*fields, FormatFieldRaw(0x0A, 0x0B, "Boranium",
		fmt.Sprintf("0x%02X%02X", d[11], d[10]),
		fmt.Sprintf("uint16 LE = %d kT", ob.Boranium)))

	// Bytes 12-13: Germanium
	*fields = append(*fields, FormatFieldRaw(0x0C, 0x0D, "Germanium",
		fmt.Sprintf("0x%02X%02X", d[13], d[12]),
		fmt.Sprintf("uint16 LE = %d kT", ob.Germanium)))

	// Bytes 14-15: wtMax|iDecayRate
	if len(d) >= 16 {
		wtMaxDecay := encoding.Read16(d, 14)
		*fields = append(*fields, FormatFieldRaw(0x0E, 0x0F, "wtMax|iDecay",
			fmt.Sprintf("0x%02X%02X", d[15], d[14]),
			fmt.Sprintf("uint16 LE = 0x%04X", wtMaxDecay)))
		*fields = append(*fields, fmt.Sprintf("           %s bits 0-13: MaxWeight = %d kT",
			TreeBranch, ob.PacketMaxWeight))
		*fields = append(*fields, fmt.Sprintf("           %s bits 14-15: DecayRate = %d",
			TreeEnd, ob.PacketDecayRate))
	}

	// Bytes 16-17: Turn number
	if len(d) >= 18 {
		*fields = append(*fields, FormatFieldRaw(0x10, 0x11, "TurnNumber",
			fmt.Sprintf("0x%02X%02X", d[17], d[16]),
			fmt.Sprintf("uint16 LE = %d", ob.PacketTurnNumber)))
	}

	// Summary
	*fields = append(*fields, "")
	*fields = append(*fields, "── Summary ──")
	if ob.IsSalvageObject {
		*fields = append(*fields, fmt.Sprintf("  Salvage #%d @ (%d, %d)", ob.Number, ob.X, ob.Y))
		*fields = append(*fields, fmt.Sprintf("  Source: Fleet #%d", ob.SourceFleetID+1))
	} else {
		*fields = append(*fields, fmt.Sprintf("  Packet #%d @ (%d, %d)", ob.Number, ob.X, ob.Y))
		*fields = append(*fields, fmt.Sprintf("  Destination: Planet #%d at Warp %d", ob.DestinationPlanetID+1, ob.WarpSpeed()))
	}
	*fields = append(*fields, fmt.Sprintf("  Owner: Player %d", ob.Owner+1))
	*fields = append(*fields, fmt.Sprintf("  Cargo: %d kT Iron, %d kT Bor, %d kT Germ (total: %d kT)",
		ob.Ironium, ob.Boranium, ob.Germanium, ob.TotalMinerals()))
}

func formatWormhole(fields *[]string, ob blocks.ObjectBlock, d []byte) {
	if len(d) < 14 {
		return
	}

	*fields = append(*fields, "")
	*fields = append(*fields, "── Wormhole Data ──")

	// Bytes 6-7: Stability/movement word
	stabilityWord := encoding.Read16(d, 6)
	*fields = append(*fields, FormatFieldRaw(0x06, 0x07, "StabilityWord",
		fmt.Sprintf("0x%02X%02X", d[7], d[6]),
		fmt.Sprintf("uint16 LE = 0x%04X", stabilityWord)))
	*fields = append(*fields, fmt.Sprintf("           %s bits 0-1: StabilityIndex = %d (%s)",
		TreeBranch, ob.StabilityIndex, ob.StabilityName()))
	*fields = append(*fields, fmt.Sprintf("           %s bits 2-11: TurnsSinceMove = %d",
		TreeBranch, ob.TurnsSinceMove))
	*fields = append(*fields, fmt.Sprintf("           %s bit 12: DestKnown = %t",
		TreeBranch, ob.DestKnown))
	*fields = append(*fields, fmt.Sprintf("           %s bit 13: IncludeInDisplay = %t",
		TreeEnd, ob.IncludeInDisplay))

	// Bytes 8-9: Can see bits (grbitPlr)
	*fields = append(*fields, FormatFieldRaw(0x08, 0x09, "CanSeeBits",
		fmt.Sprintf("0x%02X%02X", d[9], d[8]),
		fmt.Sprintf("bitmask = 0x%04X (grbitPlr)", ob.CanSeeBits)))

	// Show who can see
	var canSee []string
	for i := 0; i < 16; i++ {
		if ob.PlayerCanSee(i) {
			canSee = append(canSee, fmt.Sprintf("P%d", i+1))
		}
	}
	if len(canSee) > 0 {
		*fields = append(*fields, fmt.Sprintf("           %s Visible to: %v", TreeEnd, canSee))
	}

	// Bytes 10-11: Been through bits (grbitPlrTrav)
	*fields = append(*fields, FormatFieldRaw(0x0A, 0x0B, "BeenThroughBits",
		fmt.Sprintf("0x%02X%02X", d[11], d[10]),
		fmt.Sprintf("bitmask = 0x%04X (grbitPlrTrav)", ob.BeenThroughBits)))

	// Show who has been through
	var beenThrough []string
	for i := 0; i < 16; i++ {
		if ob.PlayerBeenThrough(i) {
			beenThrough = append(beenThrough, fmt.Sprintf("P%d", i+1))
		}
	}
	if len(beenThrough) > 0 {
		*fields = append(*fields, fmt.Sprintf("           %s Been through: %v", TreeEnd, beenThrough))
	}

	// Bytes 12-13: Target wormhole ID (idPartner)
	*fields = append(*fields, FormatFieldRaw(0x0C, 0x0D, "TargetId",
		fmt.Sprintf("0x%02X%02X", d[13], d[12]),
		fmt.Sprintf("uint16 LE = Wormhole #%d (idPartner)", ob.TargetId)))

	// Bytes 14-15: Padding
	if len(d) >= 16 {
		*fields = append(*fields, FormatFieldRaw(0x0E, 0x0F, "Padding",
			fmt.Sprintf("0x%02X%02X", d[15], d[14]),
			fmt.Sprintf("uint16 LE = 0x%04X (unused)", ob.WormholePadding)))
	}

	// Bytes 16-17: Turn number
	if len(d) >= 18 {
		*fields = append(*fields, FormatFieldRaw(0x10, 0x11, "TurnNumber",
			fmt.Sprintf("0x%02X%02X", d[17], d[16]),
			fmt.Sprintf("uint16 LE = %d", ob.WormholeTurnNumber)))
	}

	// Summary
	*fields = append(*fields, "")
	*fields = append(*fields, "── Summary ──")
	*fields = append(*fields, fmt.Sprintf("  Wormhole #%d @ (%d, %d)", ob.Number, ob.X, ob.Y))
	*fields = append(*fields, fmt.Sprintf("  Stability: %s (index %d)", ob.StabilityName(), ob.StabilityIndex))
	*fields = append(*fields, fmt.Sprintf("  Last moved: %d turns ago", ob.TurnsSinceMove))
	*fields = append(*fields, fmt.Sprintf("  Target: Wormhole #%d", ob.TargetId))
}

func formatMysteryTrader(fields *[]string, ob blocks.ObjectBlock, d []byte) {
	if len(d) < 18 {
		return
	}

	*fields = append(*fields, "")
	*fields = append(*fields, "── Mystery Trader Data ──")

	// Bytes 6-7: Destination X
	*fields = append(*fields, FormatFieldRaw(0x06, 0x07, "XDest",
		fmt.Sprintf("0x%02X%02X", d[7], d[6]),
		fmt.Sprintf("uint16 LE = %d", ob.XDest)))

	// Bytes 8-9: Destination Y
	*fields = append(*fields, FormatFieldRaw(0x08, 0x09, "YDest",
		fmt.Sprintf("0x%02X%02X", d[9], d[8]),
		fmt.Sprintf("uint16 LE = %d", ob.YDest)))

	// Byte 10: Warp
	*fields = append(*fields, FormatFieldRaw(0x0A, 0x0A, "Warp",
		fmt.Sprintf("0x%02X", d[10]),
		fmt.Sprintf("(byte & 0x0F) = Warp %d", ob.Warp)))

	// Byte 11: Unknown
	*fields = append(*fields, FormatFieldRaw(0x0B, 0x0B, "Unknown",
		fmt.Sprintf("0x%02X", d[11]),
		"TBD"))

	// Bytes 12-13: Met bits
	*fields = append(*fields, FormatFieldRaw(0x0C, 0x0D, "MetBits",
		fmt.Sprintf("0x%02X%02X", d[13], d[12]),
		fmt.Sprintf("bitmask = 0x%04X", ob.MetBits)))

	// Show who trader has met
	var metPlayers []string
	for i := 0; i < 16; i++ {
		if ob.TraderHasMet(i) {
			metPlayers = append(metPlayers, fmt.Sprintf("P%d", i+1))
		}
	}
	if len(metPlayers) > 0 {
		*fields = append(*fields, fmt.Sprintf("           %s Has met: %v", TreeEnd, metPlayers))
	}

	// Bytes 14-15: Item bits
	*fields = append(*fields, FormatFieldRaw(0x0E, 0x0F, "ItemBits",
		fmt.Sprintf("0x%02X%02X", d[15], d[14]),
		fmt.Sprintf("bitmask = 0x%04X", ob.ItemBits)))

	// Show items the trader is carrying
	traderItems := []uint16{
		blocks.TraderItemMultiCargoPod,
		blocks.TraderItemMultiFunctionPod,
		blocks.TraderItemLangstonShield,
		blocks.TraderItemMegaPolyShell,
		blocks.TraderItemAlienMiner,
		blocks.TraderItemHushABoom,
		blocks.TraderItemAntiMatterTorpedo,
		blocks.TraderItemMultiContainedMunition,
		blocks.TraderItemMiniMorph,
		blocks.TraderItemEnigmaPulsar,
		blocks.TraderItemGenesisDevice,
		blocks.TraderItemJumpGate,
		blocks.TraderItemShip,
	}

	var carriedItems []string
	for _, itemBit := range traderItems {
		if ob.TraderHasItem(itemBit) {
			carriedItems = append(carriedItems, traderItemName(itemBit))
		}
	}
	if len(carriedItems) > 0 {
		for i, item := range carriedItems {
			prefix := TreeBranch
			if i == len(carriedItems)-1 {
				prefix = TreeEnd
			}
			*fields = append(*fields, fmt.Sprintf("           %s %s", prefix, item))
		}
	}

	// Bytes 16-17: Turn number
	*fields = append(*fields, FormatFieldRaw(0x10, 0x11, "TurnNo",
		fmt.Sprintf("0x%02X%02X", d[17], d[16]),
		fmt.Sprintf("uint16 LE = %d", ob.TurnNo)))

	// Summary
	*fields = append(*fields, "")
	*fields = append(*fields, "── Summary ──")
	*fields = append(*fields, fmt.Sprintf("  Mystery Trader #%d @ (%d, %d)", ob.Number, ob.X, ob.Y))
	*fields = append(*fields, fmt.Sprintf("  Heading to: (%d, %d) at Warp %d", ob.XDest, ob.YDest, ob.Warp))
	*fields = append(*fields, fmt.Sprintf("  Items: %d", len(carriedItems)))
	*fields = append(*fields, fmt.Sprintf("  Met players: %d", len(metPlayers)))
}
