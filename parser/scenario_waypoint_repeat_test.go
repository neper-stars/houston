package parser_test

import (
	"os"
	"testing"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/parser"
)

func TestScenarioWaypointRepeat(t *testing.T) {
	data, err := os.ReadFile("../testdata/scenario-waypoint-repeat/game.m1")
	if err != nil {
		t.Fatalf("Failed to read M file: %v", err)
	}

	fd := parser.FileData(data)
	blockList, err := fd.BlockList()
	if err != nil {
		t.Fatalf("Failed to parse blocks: %v", err)
	}

	// Find Fleet 11 (Rubber Pump)
	var rubberPump *blocks.FleetBlock
	var waypointTasks []blocks.WaypointTaskBlock

	for _, block := range blockList {
		if fb, ok := block.(blocks.FleetBlock); ok && fb.FleetNumber == 11 {
			rubberPump = &fb
		}
		if wtb, ok := block.(blocks.WaypointTaskBlock); ok {
			waypointTasks = append(waypointTasks, wtb)
		}
	}

	if rubberPump == nil {
		t.Fatal("Fleet 11 (Rubber Pump) not found")
	}

	// Verify fleet properties
	t.Run("FleetProperties", func(t *testing.T) {
		if rubberPump.X != 1267 {
			t.Errorf("Expected X=1267, got %d", rubberPump.X)
		}
		if rubberPump.Y != 1327 {
			t.Errorf("Expected Y=1327, got %d", rubberPump.Y)
		}
		if !rubberPump.RepeatOrders {
			t.Error("Expected RepeatOrders=true")
		}
		if rubberPump.WaypointCount != 3 {
			t.Errorf("Expected WaypointCount=3, got %d", rubberPump.WaypointCount)
		}
	})

	// Find waypoints for Rubber Pump (to Hurl and Rubber)
	var hurlWaypoint, rubberWaypoint *blocks.WaypointTaskBlock
	for i := range waypointTasks {
		wtb := &waypointTasks[i]
		// Hurl is at X=1273, Y=1341
		if wtb.X == 1273 && wtb.Y == 1341 && wtb.WaypointTask == blocks.WaypointTaskTransport {
			hurlWaypoint = wtb
		}
		// Rubber is at X=1249, Y=1149
		if wtb.X == 1249 && wtb.Y == 1149 && wtb.WaypointTask == blocks.WaypointTaskTransport {
			rubberWaypoint = wtb
		}
	}

	t.Run("HurlWaypoint_LoadAllAvailable", func(t *testing.T) {
		if hurlWaypoint == nil {
			t.Fatal("Waypoint to Hurl not found")
		}
		if hurlWaypoint.WaypointTask != blocks.WaypointTaskTransport {
			t.Errorf("Expected task=Transport(1), got %d", hurlWaypoint.WaypointTask)
		}
		if !hurlWaypoint.IsLoadAllTransport() {
			t.Error("Expected Load All Available transport action")
		}
		if hurlWaypoint.TransportAction != blocks.TransportActionLoadAll {
			t.Errorf("Expected TransportAction=0x10, got 0x%02X", hurlWaypoint.TransportAction)
		}
	})

	t.Run("RubberWaypoint_UnloadAll", func(t *testing.T) {
		if rubberWaypoint == nil {
			t.Fatal("Waypoint to Rubber not found")
		}
		if rubberWaypoint.WaypointTask != blocks.WaypointTaskTransport {
			t.Errorf("Expected task=Transport(1), got %d", rubberWaypoint.WaypointTask)
		}
		if !rubberWaypoint.IsUnloadAllTransport() {
			t.Error("Expected Unload All transport action")
		}
		if rubberWaypoint.TransportAction != blocks.TransportActionUnloadAll {
			t.Errorf("Expected TransportAction=0x20, got 0x%02X", rubberWaypoint.TransportAction)
		}
	})
}
