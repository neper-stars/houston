package blockdetail

import (
	"fmt"

	"github.com/neper-stars/houston/blocks"
	"github.com/neper-stars/houston/encoding"
)

func init() {
	RegisterFormatter(blocks.MessageBlockType, FormatMessage)
}

// FormatMessage provides detailed view for MessageBlock (type 40)
func FormatMessage(block blocks.Block, index int) string {
	width := DefaultWidth
	mb, ok := block.(blocks.MessageBlock)
	if !ok {
		return FormatGeneric(block, index)
	}

	d := mb.DecryptedData()
	header := FormatBlockHeader(block, index, width)
	hexSection := FormatHexSection(d, width)

	var fields []string

	if len(d) < 10 {
		fields = append(fields, "(block too short)")
		fieldsSection := FormatFieldsSection(fields, width)
		return BuildOutput(header, hexSection, fieldsSection)
	}

	// Bytes 0-1: Unknown word 0
	word0 := encoding.Read16(d, 0)
	fields = append(fields, FormatFieldRaw(0x00, 0x01, "UnknownWord0",
		fmt.Sprintf("0x%02X%02X", d[1], d[0]),
		fmt.Sprintf("uint16 LE = %d (0x%04X)", word0, word0)))

	// Bytes 2-3: Unknown word 2
	word2 := encoding.Read16(d, 2)
	fields = append(fields, FormatFieldRaw(0x02, 0x03, "UnknownWord2",
		fmt.Sprintf("0x%02X%02X", d[3], d[2]),
		fmt.Sprintf("uint16 LE = %d (0x%04X)", word2, word2)))

	// Bytes 4-5: Sender ID
	senderId := encoding.Read16(d, 4)
	senderDisplay := int(senderId) + 1
	fields = append(fields, FormatFieldRaw(0x04, 0x05, "SenderId",
		fmt.Sprintf("0x%02X%02X", d[5], d[4]),
		fmt.Sprintf("uint16 LE = %d -> Player %d", senderId, senderDisplay)))

	// Bytes 6-7: Receiver ID
	receiverId := encoding.Read16(d, 6)
	receiverStr := fmt.Sprintf("Player %d", receiverId)
	if receiverId == 0 {
		receiverStr = "Everyone (broadcast)"
	}
	fields = append(fields, FormatFieldRaw(0x06, 0x07, "ReceiverId",
		fmt.Sprintf("0x%02X%02X", d[7], d[6]),
		fmt.Sprintf("uint16 LE = %d -> %s", receiverId, receiverStr)))

	// Bytes 8-9: Unknown word 8 (message type indicator)
	word8 := encoding.Read16(d, 8)
	word8Str := fmt.Sprintf("%d", word8)
	switch word8 {
	case 3:
		word8Str = "3 (reply)"
	case 4:
		word8Str = "4 (normal)"
	}
	fields = append(fields, FormatFieldRaw(0x08, 0x09, "UnknownWord8",
		fmt.Sprintf("0x%02X%02X", d[9], d[8]),
		fmt.Sprintf("uint16 LE = %s", word8Str)))

	// Message data section
	if len(d) > 10 {
		fields = append(fields, "")
		fields = append(fields, "── Message Data ──")

		messageData := d[10:]
		if len(messageData) >= 2 {
			// Message header
			msgHeader := encoding.Read16(messageData, 0)
			byteSize := int(msgHeader & 0x3FF)
			asciiIndicator := int(msgHeader >> 10)

			fields = append(fields, FormatFieldRaw(0x0A, 0x0B, "MsgHeader",
				fmt.Sprintf("0x%02X%02X", messageData[1], messageData[0]),
				fmt.Sprintf("uint16 LE = 0x%04X", msgHeader)))

			useAscii := asciiIndicator == 0x3F
			actualByteSize := byteSize
			if useAscii {
				actualByteSize = (^byteSize) & 0x3FF
			}

			fields = append(fields, fmt.Sprintf("           %s byteSize = (header & 0x3FF) = %d", TreeBranch, byteSize))
			fields = append(fields, fmt.Sprintf("           %s asciiIndicator = (header >> 10) = 0x%02X", TreeBranch, asciiIndicator))
			if useAscii {
				fields = append(fields, fmt.Sprintf("           %s ASCII mode: actualSize = (~%d) & 0x3FF = %d", TreeBranch, byteSize, actualByteSize))
			}
			fields = append(fields, fmt.Sprintf("           %s encoding = %s", TreeEnd, map[bool]string{true: "ASCII", false: "Stars! encoded"}[useAscii]))

			// Message text bytes
			if len(messageData) > 2 {
				textStart := 0x0C
				textEnd := 0x0A + len(messageData) - 1
				fields = append(fields, "")
				fields = append(fields, fmt.Sprintf("0x%02X-0x%02X: Message Text (%d bytes)",
					textStart, textEnd, len(messageData)-2))
			}
		}

		// Decoded message
		fields = append(fields, "")
		fields = append(fields, "── Decoded Message ──")
		if mb.Message != "" {
			fields = append(fields, fmt.Sprintf("  %q", mb.Message))
		} else {
			fields = append(fields, "  (empty)")
		}

		// Message metadata
		fields = append(fields, "")
		fields = append(fields, "── Summary ──")
		fields = append(fields, fmt.Sprintf("  From: Player %d", mb.SenderDisplayId()))
		if mb.IsBroadcast() {
			fields = append(fields, "  To: Everyone (broadcast)")
		} else {
			fields = append(fields, fmt.Sprintf("  To: Player %d", mb.ReceiverDisplayId()))
		}
		if mb.IsReply() {
			fields = append(fields, "  Type: Reply")
		} else {
			fields = append(fields, "  Type: Normal")
		}
	}

	fieldsSection := FormatFieldsSection(fields, width)
	return BuildOutput(header, hexSection, fieldsSection)
}
