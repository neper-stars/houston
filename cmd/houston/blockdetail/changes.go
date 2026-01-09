package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.ResearchChangeBlockType, FormatResearchChange)
	RegisterFormatter(blocks.PlanetChangeBlockType, FormatPlanetChange)
	RegisterFormatter(blocks.ChangePasswordBlockType, FormatChangePassword)
}

// FormatResearchChange provides detailed view for ResearchChangeBlock (type 34)
func FormatResearchChange(block blocks.Block, index int) string {
	width := DefaultWidth
	rcb, ok := block.(blocks.ResearchChangeBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := rcb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 2 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Byte 0: Budget percentage
	fields = append(fields, FormatFieldRaw(0x00, 0x00, "BudgetPercent",
		fmt.Sprintf("0x%02X", d[0]),
		fmt.Sprintf("%d%%", rcb.BudgetPercent)))

	// Byte 1: Current and next field
	fields = append(fields, FormatFieldRaw(0x01, 0x01, "Fields",
		fmt.Sprintf("0x%02X", d[1]),
		fmt.Sprintf("0b%08b", d[1])))
	fields = append(fields, fmt.Sprintf("           %s bits0-3: currentField = %d (%s)",
		TreeBranch, rcb.CurrentField, blocks.ResearchFieldName(rcb.CurrentField)))
	fields = append(fields, fmt.Sprintf("           %s bits4-7: nextField = %d (%s)",
		TreeEnd, rcb.NextField, blocks.ResearchFieldName(rcb.NextField)))

	// Additional bytes if present
	if len(d) > 2 {
		fields = append(fields, "")
		fields = append(fields, "── Additional Data ──")
		for i := 2; i < len(d); i++ {
			fields = append(fields, FormatFieldRaw(i, i, fmt.Sprintf("Byte%d", i),
				fmt.Sprintf("0x%02X", d[i]),
				"TBD"))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Research budget: %d%%", rcb.BudgetPercent))
	fields = append(fields, fmt.Sprintf("  Current field: %s", blocks.ResearchFieldName(rcb.CurrentField)))
	fields = append(fields, fmt.Sprintf("  Next field: %s", blocks.ResearchFieldName(rcb.NextField)))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatPlanetChange provides detailed view for PlanetChangeBlock (type 35)
func FormatPlanetChange(block blocks.Block, index int) string {
	width := DefaultWidth
	pcb, ok := block.(blocks.PlanetChangeBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := pcb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 6 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Planet ID (uint16 LE)
	planetWord := encoding.Read16(d, 0)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "PlanetId",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = %d -> Planet #%d", planetWord, pcb.PlanetId+1)))

	// Bytes 2-3: Packed settings (uint16 LE)
	settingsWord := encoding.Read16(d, 2)
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "Settings",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = 0x%04X (0b%016b)", settingsWord, settingsWord)))
	contributeStr := "NO"
	if pcb.ContributeLeftover {
		contributeStr = "YES"
	}
	fields = append(fields, fmt.Sprintf("           %s bit 0:      fNoResearch = %s (contribute leftover to research)",
		TreeBranch, contributeStr))
	fields = append(fields, fmt.Sprintf("           %s bits 1-10:  pctDp = %d (driver/packet %%)",
		TreeBranch, pcb.DriverPacketPercent))
	fields = append(fields, fmt.Sprintf("           %s bits 11-14: iWarpFling = %d (packet warp speed)",
		TreeEnd, pcb.PacketWarpSpeed))

	// Bytes 4-5: Route destination (uint16 LE, left-shifted by 1)
	routeWord := encoding.Read16(d, 4)
	routeDesc := "no route"
	if pcb.RouteDestinationId > 0 {
		routeDesc = fmt.Sprintf("-> Planet #%d", pcb.RouteDestinationId+1)
	}
	fields = append(fields, FormatFieldRaw(0x04, 0x05, "RouteDest",
		fmt.Sprintf("0x%02X%02X", d[5], d[4]),
		fmt.Sprintf("uint16 LE = %d (raw), >> 1 = %d (%s)", routeWord, pcb.RouteDestinationId, routeDesc)))

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Planet #%d settings change:", pcb.PlanetId+1))
	fields = append(fields, fmt.Sprintf("    %s Contribute leftover to research: %s", TreeBranch, contributeStr))
	fields = append(fields, fmt.Sprintf("    %s Driver/Packet %%: %d%%", TreeBranch, pcb.DriverPacketPercent))
	fields = append(fields, fmt.Sprintf("    %s Packet Warp Speed: %d", TreeBranch, pcb.PacketWarpSpeed))
	if pcb.RouteDestinationId > 0 {
		fields = append(fields, fmt.Sprintf("    %s Route Destination: Planet #%d", TreeEnd, pcb.RouteDestinationId+1))
	} else {
		fields = append(fields, fmt.Sprintf("    %s Route Destination: None", TreeEnd))
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatChangePassword provides detailed view for ChangePasswordBlock (type 36)
func FormatChangePassword(block blocks.Block, index int) string {
	width := DefaultWidth
	cpb, ok := block.(blocks.ChangePasswordBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := cpb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 4 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-3: New password hash
	fields = append(fields, FormatFieldRaw(0x00, 0x03, "NewPasswordHash",
		fmt.Sprintf("0x%02X%02X%02X%02X", d[3], d[2], d[1], d[0]),
		fmt.Sprintf("uint32 LE = 0x%08X", cpb.NewPasswordHash)))

	// Interpret hash
	if cpb.HasPassword() {
		fields = append(fields, fmt.Sprintf("           %s Hash is non-zero: setting new password", TreeEnd))
	} else {
		fields = append(fields, fmt.Sprintf("           %s Hash is zero: removing password", TreeEnd))
	}

	// Additional bytes if present
	if len(d) > 4 {
		fields = append(fields, "")
		fields = append(fields, "── Additional Data ──")
		for i := 4; i < len(d); i++ {
			fields = append(fields, FormatFieldRaw(i, i, fmt.Sprintf("Byte%d", i),
				fmt.Sprintf("0x%02X", d[i]),
				"TBD"))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	if cpb.HasPassword() {
		fields = append(fields, fmt.Sprintf("  Action: Set new password (hash=0x%08X)", cpb.NewPasswordHash))
	} else {
		fields = append(fields, "  Action: Remove password")
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
