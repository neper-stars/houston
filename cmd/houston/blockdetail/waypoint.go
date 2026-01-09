package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
)

func init() {
	RegisterFormatter(blocks.WaypointDeleteBlockType, FormatWaypointDelete)
	RegisterFormatter(blocks.WaypointAddBlockType, FormatWaypointAdd)
	RegisterFormatter(blocks.WaypointChangeTaskBlockType, FormatWaypointChangeTask)
	RegisterFormatter(blocks.WaypointRepeatOrdersBlockType, FormatWaypointRepeatOrders)
	RegisterFormatter(blocks.WaypointTaskTypeChangeBlockType, FormatWaypointTaskTypeChange)
	RegisterFormatter(blocks.WaypointTaskBlockType, FormatWaypointTask)
	RegisterFormatter(blocks.WaypointBlockType, FormatWaypoint)
}

// waypointTaskName returns human-readable name for waypoint task type
// Uses the shared function from blocks package
func waypointTaskName(task int) string {
	return blocks.WaypointTaskName(task)
}

// waypointTargetTypeName returns human-readable name for target type
func waypointTargetTypeName(targetType int) string {
	names := map[int]string{
		blocks.WaypointTargetPlanet:    "Planet",
		blocks.WaypointTargetFleet:     "Fleet",
		blocks.WaypointTargetDeepSpace: "Deep Space",
		blocks.WaypointTargetWormhole:  "Wormhole/MysteryTrader/Minefield",
	}
	if name, ok := names[targetType]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", targetType)
}

// formatWarpSpeed formats warp speed including stargate
func formatWarpSpeed(warp int) string {
	if warp == blocks.WarpStargate {
		return "Stargate"
	}
	return fmt.Sprintf("Warp %d", warp)
}

// FormatWaypointDelete provides detailed view for WaypointDeleteBlock (type 3)
func FormatWaypointDelete(block blocks.Block, index int) string {
	width := DefaultWidth
	wdb, ok := block.(blocks.WaypointDeleteBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := wdb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 3 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Fleet number (9 bits)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "FleetNumber",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("(d[0] + (d[1]&0x01)<<8) = %d -> Fleet #%d", wdb.FleetNumber, wdb.FleetNumber+1)))

	// Byte 2: Waypoint number
	fields = append(fields, FormatFieldRaw(0x02, 0x02, "WaypointNumber",
		fmt.Sprintf("0x%02X", d[2]),
		fmt.Sprintf("%d", wdb.WaypointNumber)))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatWaypointAdd provides detailed view for WaypointAddBlock (type 4)
func FormatWaypointAdd(block blocks.Block, index int) string {
	width := DefaultWidth
	wab, ok := block.(blocks.WaypointAddBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	return formatWaypointChangeTaskCommon(&wab.WaypointChangeTaskBlock, block, index, width)
}

// FormatWaypointChangeTask provides detailed view for WaypointChangeTaskBlock (type 5)
func FormatWaypointChangeTask(block blocks.Block, index int) string {
	width := DefaultWidth
	wctb, ok := block.(blocks.WaypointChangeTaskBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	return formatWaypointChangeTaskCommon(&wctb, block, index, width)
}

// formatWaypointChangeTaskCommon handles common formatting for WaypointAddBlock and WaypointChangeTaskBlock
func formatWaypointChangeTaskCommon(wctb *blocks.WaypointChangeTaskBlock, block blocks.Block, index int, width int) string {
	d := wctb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 12 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Fleet number (9 bits)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "FleetNumber",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("(d[0] + (d[1]&0x01)<<8) = %d -> Fleet #%d", wctb.FleetNumber, wctb.FleetNumber+1)))

	// Bytes 2-3: Waypoint index (uint16 LE)
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "WaypointIndex",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = %d", wctb.WaypointIndex)))

	// Bytes 4-5: X coordinate
	fields = append(fields, FormatFieldRaw(0x04, 0x05, "X",
		fmt.Sprintf("0x%02X%02X", d[5], d[4]),
		fmt.Sprintf("uint16 LE = %d", wctb.X)))

	// Bytes 6-7: Y coordinate
	fields = append(fields, FormatFieldRaw(0x06, 0x07, "Y",
		fmt.Sprintf("0x%02X%02X", d[7], d[6]),
		fmt.Sprintf("uint16 LE = %d", wctb.Y)))

	// Bytes 8-9: Target (9 bits)
	fields = append(fields, FormatFieldRaw(0x08, 0x09, "Target",
		fmt.Sprintf("0x%02X%02X", d[9], d[8]),
		fmt.Sprintf("(d[0] + (d[1]&0x01)<<8) = %d", wctb.Target)))

	// Byte 10: Warp (upper nibble) | WaypointTask (lower nibble)
	fields = append(fields, FormatFieldRaw(0x0A, 0x0A, "Warp/Task",
		fmt.Sprintf("0x%02X", d[10]),
		fmt.Sprintf("warp=(d>>4)=%d (%s), task=(d&0x0F)=%d (%s)",
			wctb.Warp, formatWarpSpeed(wctb.Warp), wctb.WaypointTask, waypointTaskName(wctb.WaypointTask))))

	// Byte 11: flags (bit4=fValidTask, bit5=fNoAutoTrack) | TargetType (lower nibble)
	validTaskStr := "false"
	if wctb.ValidTask {
		validTaskStr = "true"
	}
	noAutoTrackStr := "false"
	if wctb.NoAutoTrack {
		noAutoTrackStr = "true"
	}
	fields = append(fields, FormatFieldRaw(0x0B, 0x0B, "Flags/TargetType",
		fmt.Sprintf("0x%02X", d[11]),
		fmt.Sprintf("targetType=(d&0x0F)=%d (%s), fValidTask=%s, fNoAutoTrack=%s",
			wctb.TargetType, waypointTargetTypeName(wctb.TargetType), validTaskStr, noAutoTrackStr)))

	// Task-specific data
	switch {
	case wctb.WaypointTask == blocks.WaypointTaskTransport && len(d) >= 22:
		fields = append(fields, "")
		fields = append(fields, "── Transport Orders ──")
		formatTransportOrders(&fields, d, 12, wctb.TransportOrders[:])
	case wctb.WaypointTask == blocks.WaypointTaskPatrol && len(d) >= 15:
		fields = append(fields, "")
		fields = append(fields, "── Patrol Data ──")
		fields = append(fields, FormatFieldRaw(0x0C, 0x0C, "SubTaskIndex",
			fmt.Sprintf("0x%02X", d[12]),
			fmt.Sprintf("%d", wctb.SubTaskIndex)))
		fields = append(fields, FormatFieldRaw(0x0D, 0x0D, "Unknown",
			fmt.Sprintf("0x%02X", d[13]),
			"TBD"))
		fields = append(fields, FormatFieldRaw(0x0E, 0x0E, "PatrolRange",
			fmt.Sprintf("0x%02X", d[14]),
			fmt.Sprintf("%d -> %s", wctb.PatrolRange, blocks.PatrolRangeName(wctb.PatrolRange))))
	case len(d) > 12:
		fields = append(fields, "")
		fields = append(fields, FormatFieldRaw(0x0C, 0x0C, "SubTaskIndex",
			fmt.Sprintf("0x%02X", d[12]),
			fmt.Sprintf("%d", wctb.SubTaskIndex)))
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatWaypointRepeatOrders provides detailed view for WaypointRepeatOrdersBlock (type 10)
// See reversing_notes/BT-10-WaypointRepeatOrders.md for format details.
func FormatWaypointRepeatOrders(block blocks.Block, index int) string {
	width := DefaultWidth
	wrob, ok := block.(blocks.WaypointRepeatOrdersBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := wrob.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 3 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Fleet number (9 bits)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "FleetNumber",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("(d[0] + (d[1]&0x01)<<8) = %d -> Fleet #%d", wrob.FleetNumber, wrob.FleetNumber+1)))

	// Byte 2: Packed byte - bit 0 = EnableRepeat, bits 1-7 = RepeatFromWaypoint
	fields = append(fields, FormatFieldRaw(0x02, 0x02, "EnableRepeat+Waypoint",
		fmt.Sprintf("0x%02X", d[2]),
		fmt.Sprintf("0b%08b", d[2])))
	fields = append(fields, fmt.Sprintf("           %s bit0: EnableRepeat = %v", TreeBranch, wrob.EnableRepeat))
	fields = append(fields, fmt.Sprintf("           %s bits1-7: RepeatFromWaypoint = %d", TreeEnd, wrob.RepeatFromWaypoint))

	// Byte 3: Padding
	if len(d) >= 4 {
		fields = append(fields, FormatFieldRaw(0x03, 0x03, "Padding",
			fmt.Sprintf("0x%02X", d[3]),
			"(unused)"))
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	if wrob.EnableRepeat {
		fields = append(fields, fmt.Sprintf("  Fleet #%d: ENABLE repeat orders from waypoint %d",
			wrob.FleetNumber+1, wrob.RepeatFromWaypoint))
	} else {
		fields = append(fields, fmt.Sprintf("  Fleet #%d: DISABLE repeat orders",
			wrob.FleetNumber+1))
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatWaypointTaskTypeChange provides detailed view for WaypointTaskTypeChangeBlock (type 11)
func FormatWaypointTaskTypeChange(block blocks.Block, index int) string {
	width := DefaultWidth
	wttcb, ok := block.(blocks.WaypointTaskTypeChangeBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := wttcb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 6 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Fleet ID
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "FleetID",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = %d -> Fleet #%d", wttcb.FleetID, wttcb.FleetID+1)))

	// Bytes 2-3: Waypoint index
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "WaypointIndex",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = %d (0-based index)", wttcb.WaypointIndex)))

	// Bytes 4-5: Task type
	fields = append(fields, FormatFieldRaw(0x04, 0x05, "TaskType",
		fmt.Sprintf("0x%02X%02X", d[5], d[4]),
		fmt.Sprintf("uint16 LE = %d (%s)", wttcb.TaskType, wttcb.TaskTypeName())))

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Fleet #%d, Waypoint %d -> Set task to %s",
		wttcb.FleetID+1, wttcb.WaypointIndex, wttcb.TaskTypeName()))

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// FormatWaypointTask provides detailed view for WaypointTaskBlock (type 19)
func FormatWaypointTask(block blocks.Block, index int) string {
	width := DefaultWidth
	wtb, ok := block.(blocks.WaypointTaskBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	return formatWaypointCommon(&wtb.WaypointBlock, block, index, width)
}

// FormatWaypoint provides detailed view for WaypointBlock (type 20)
func FormatWaypoint(block blocks.Block, index int) string {
	width := DefaultWidth
	wb, ok := block.(blocks.WaypointBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	return formatWaypointCommon(&wb, block, index, width)
}

// formatWaypointCommon handles common formatting for WaypointBlock and WaypointTaskBlock
func formatWaypointCommon(wb *blocks.WaypointBlock, block blocks.Block, index int, width int) string {
	d := wb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 8 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: X coordinate
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "X",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = %d", wb.X)))

	// Bytes 2-3: Y coordinate
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "Y",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = %d", wb.Y)))

	// Bytes 4-5: Position object ID
	fields = append(fields, FormatFieldRaw(0x04, 0x05, "PositionObject",
		fmt.Sprintf("0x%02X%02X", d[5], d[4]),
		fmt.Sprintf("uint16 LE = %d", wb.PositionObject)))

	// Byte 6: Warp (upper nibble) | WaypointTask (lower nibble)
	fields = append(fields, FormatFieldRaw(0x06, 0x06, "Warp/Task",
		fmt.Sprintf("0x%02X", d[6]),
		fmt.Sprintf("warp=(d>>4)=%d (%s), task=(d&0x0F)=%d (%s)",
			wb.Warp, formatWarpSpeed(wb.Warp), wb.WaypointTask, waypointTaskName(wb.WaypointTask))))

	// Byte 7: Position object type
	fields = append(fields, FormatFieldRaw(0x07, 0x07, "PositionObjectType",
		fmt.Sprintf("0x%02X", d[7]),
		fmt.Sprintf("%d (%s)", wb.PositionObjectType, waypointTargetTypeName(wb.PositionObjectType))))

	// Task-specific data
	switch {
	case wb.WaypointTask == blocks.WaypointTaskTransport && len(d) >= 18:
		fields = append(fields, "")
		fields = append(fields, "── Transport Orders (5 types: Ironium, Boranium, Germanium, Colonists, Fuel) ──")
		formatTransportOrders(&fields, d, 8, wb.TransportOrders[:])
	case wb.WaypointTask == blocks.WaypointTaskPatrol && len(d) > 8:
		fields = append(fields, "")
		fields = append(fields, "── Patrol Data ──")
		fields = append(fields, FormatFieldRaw(0x08, 0x08, "PatrolRange",
			fmt.Sprintf("0x%02X", d[8]),
			fmt.Sprintf("%d -> %s", wb.PatrolRange, blocks.PatrolRangeName(wb.PatrolRange))))
	case len(wb.AdditionalBytes) > 0:
		fields = append(fields, "")
		fields = append(fields, fmt.Sprintf("0x08-0x%02X: AdditionalBytes (%d bytes)", 7+len(wb.AdditionalBytes), len(wb.AdditionalBytes)))
		for i, b := range wb.AdditionalBytes {
			prefix := TreeBranch
			if i == len(wb.AdditionalBytes)-1 {
				prefix = TreeEnd
			}
			fields = append(fields, fmt.Sprintf("           %s [0x%02X] = 0x%02X", prefix, 8+i, b))
		}
	}

	// Summary
	fields = append(fields, "")
	fields = append(fields, "── Summary ──")
	fields = append(fields, fmt.Sprintf("  Position: (%d, %d)", wb.X, wb.Y))
	fields = append(fields, fmt.Sprintf("  Speed: %s", formatWarpSpeed(wb.Warp)))
	fields = append(fields, fmt.Sprintf("  Task: %s", waypointTaskName(wb.WaypointTask)))
	if wb.WaypointTask == blocks.WaypointTaskTransport && wb.HasTransportOrders() {
		for i := 0; i < blocks.TransportCargoTypeCount; i++ {
			if wb.TransportOrders[i].Action != blocks.TransportTaskNoAction {
				fields = append(fields, fmt.Sprintf("    %s", wb.TransportOrderDescription(i)))
			}
		}
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}

// formatTransportOrders formats transport order bytes (5 types including Fuel)
func formatTransportOrders(fields *[]string, d []byte, startOffset int, orders []blocks.TransportOrder) {
	for i := 0; i < blocks.TransportCargoTypeCount; i++ {
		offset := startOffset + (i * 2)
		if offset+1 >= len(d) {
			break
		}

		valueByte := d[offset]
		actionByte := d[offset+1]
		action := int(actionByte >> 4)
		value := int(valueByte)

		cargoName := blocks.CargoTypeName(i)
		unit := blocks.CargoTypeUnit(i)

		*fields = append(*fields, FormatFieldRaw(offset, offset+1, cargoName,
			fmt.Sprintf("0x%02X%02X", actionByte, valueByte),
			fmt.Sprintf("value=%d %s, action=%d (%s)",
				value, unit, action, blocks.TransportTaskName(action))))
	}
}
