package blocks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWaypointTaskTypeChangeBlock_Decode(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		wantFleetID  int
		wantWpIndex  int
		wantTaskType int
		wantTaskName string
	}{
		{
			name:         "set task to None",
			data:         []byte{0x05, 0x00, 0x02, 0x00, 0x00, 0x00}, // FleetID=5, WpIndex=2, Task=0
			wantFleetID:  5,
			wantWpIndex:  2,
			wantTaskType: WaypointTaskNone,
			wantTaskName: "None",
		},
		{
			name:         "set task to Transport",
			data:         []byte{0x0A, 0x00, 0x01, 0x00, 0x01, 0x00}, // FleetID=10, WpIndex=1, Task=1
			wantFleetID:  10,
			wantWpIndex:  1,
			wantTaskType: WaypointTaskTransport,
			wantTaskName: "Transport",
		},
		{
			name:         "set task to Colonize",
			data:         []byte{0x00, 0x01, 0x00, 0x00, 0x02, 0x00}, // FleetID=256, WpIndex=0, Task=2
			wantFleetID:  256,
			wantWpIndex:  0,
			wantTaskType: WaypointTaskColonize,
			wantTaskName: "Colonize",
		},
		{
			name:         "set task to Patrol",
			data:         []byte{0x03, 0x00, 0x05, 0x00, 0x07, 0x00}, // FleetID=3, WpIndex=5, Task=7
			wantFleetID:  3,
			wantWpIndex:  5,
			wantTaskType: WaypointTaskPatrol,
			wantTaskName: "Patrol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gb := GenericBlock{
				Type:      WaypointTaskTypeChangeBlockType,
				Size:      BlockSize(len(tt.data)),
				Data:      tt.data,
				Decrypted: tt.data,
			}

			block := NewWaypointTaskTypeChangeBlock(gb)

			assert.Equal(t, tt.wantFleetID, block.FleetID, "FleetID mismatch")
			assert.Equal(t, tt.wantWpIndex, block.WaypointIndex, "WaypointIndex mismatch")
			assert.Equal(t, tt.wantTaskType, block.TaskType, "TaskType mismatch")
			assert.Equal(t, tt.wantTaskName, block.TaskTypeName(), "TaskTypeName mismatch")
		})
	}
}

func TestWaypointTaskTypeChangeBlock_Encode(t *testing.T) {
	tests := []struct {
		name      string
		fleetID   int
		wpIndex   int
		taskType  int
		wantBytes []byte
	}{
		{
			name:      "encode task None",
			fleetID:   5,
			wpIndex:   2,
			taskType:  WaypointTaskNone,
			wantBytes: []byte{0x05, 0x00, 0x02, 0x00, 0x00, 0x00},
		},
		{
			name:      "encode task Transport",
			fleetID:   10,
			wpIndex:   1,
			taskType:  WaypointTaskTransport,
			wantBytes: []byte{0x0A, 0x00, 0x01, 0x00, 0x01, 0x00},
		},
		{
			name:      "encode high fleet ID",
			fleetID:   256,
			wpIndex:   0,
			taskType:  WaypointTaskColonize,
			wantBytes: []byte{0x00, 0x01, 0x00, 0x00, 0x02, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := &WaypointTaskTypeChangeBlock{
				FleetID:       tt.fleetID,
				WaypointIndex: tt.wpIndex,
				TaskType:      tt.taskType,
			}

			encoded := block.Encode()
			require.Len(t, encoded, 6, "encoded length should be 6")
			assert.Equal(t, tt.wantBytes, encoded, "encoded bytes mismatch")
		})
	}
}

func TestWaypointTaskTypeChangeBlock_RoundTrip(t *testing.T) {
	original := &WaypointTaskTypeChangeBlock{
		FleetID:       42,
		WaypointIndex: 3,
		TaskType:      WaypointTaskLayMines,
	}

	encoded := original.Encode()

	gb := GenericBlock{
		Type:      WaypointTaskTypeChangeBlockType,
		Size:      BlockSize(len(encoded)),
		Data:      encoded,
		Decrypted: encoded,
	}

	decoded := NewWaypointTaskTypeChangeBlock(gb)

	assert.Equal(t, original.FleetID, decoded.FleetID, "FleetID mismatch after round-trip")
	assert.Equal(t, original.WaypointIndex, decoded.WaypointIndex, "WaypointIndex mismatch after round-trip")
	assert.Equal(t, original.TaskType, decoded.TaskType, "TaskType mismatch after round-trip")
}

func TestWaypointTaskName(t *testing.T) {
	tests := []struct {
		task int
		want string
	}{
		{WaypointTaskNone, "None"},
		{WaypointTaskTransport, "Transport"},
		{WaypointTaskColonize, "Colonize"},
		{WaypointTaskRemoteMining, "Remote Mining"},
		{WaypointTaskMergeFleet, "Merge Fleet"},
		{WaypointTaskScrapFleet, "Scrap Fleet"},
		{WaypointTaskLayMines, "Lay Mines"},
		{WaypointTaskPatrol, "Patrol"},
		{WaypointTaskRoute, "Route"},
		{WaypointTaskTransfer, "Transfer"},
		{99, "Unknown(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := WaypointTaskName(tt.task)
			assert.Equal(t, tt.want, got)
		})
	}
}
