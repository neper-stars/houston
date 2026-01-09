package blocks

import (
	"github.com/neper-stars/houston/encoding"
)

// MessageBlock represents a player message (Type 40)
//
// The file format is the in-memory MSGPLR structure written directly to disk.
//
// Format:
//
//	Bytes 0-3:  Garbage (linked list pointer from memory - ignore)
//	Bytes 4-5:  Sender ID (iPlrFrom) - sender player index (0-15)
//	Bytes 6-7:  Recipient ID (iPlrTo) - 0=broadcast, 1-16=specific player
//	Bytes 8-9:  InReplyTo (iInRe) - message ID being replied to (0=not a reply)
//	Bytes 10-11: Length (cLen) - message byte count (negative=ASCII)
//	Bytes 12+:  Message (rgbMsg) - Stars! encoded message string
type MessageBlock struct {
	GenericBlock

	SenderId   int    // Sender player index (0-15)
	ReceiverId int    // Receiver: 0=broadcast, 1-16=specific player
	InReplyTo  int    // Message ID being replied to for threading (0=not a reply)
	Message    string // Message text
}

// NewMessageBlock creates a MessageBlock from a GenericBlock
func NewMessageBlock(b GenericBlock) *MessageBlock {
	mb := &MessageBlock{
		GenericBlock: b,
	}
	mb.decode()
	return mb
}

func (mb *MessageBlock) decode() {
	data := mb.Decrypted
	if len(data) < 12 {
		return
	}

	// Bytes 0-3 are garbage (linked list pointer) - ignore
	mb.SenderId = int(encoding.Read16(data, 4))
	mb.ReceiverId = int(encoding.Read16(data, 6))
	mb.InReplyTo = int(encoding.Read16(data, 8))

	// Bytes 10-11 contain the length header, bytes 12+ contain the message
	// decodeStarsMessage expects the length header at the start
	if len(data) > 10 {
		messageData := data[10:]
		mb.Message = decodeStarsMessage(messageData)
	}
}

// decodeStarsMessage decodes a Stars! encoded message
func decodeStarsMessage(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	header := encoding.Read16(data, 0)
	byteSize := int(header & 0x3FF)     // Lower 10 bits
	asciiIndicator := int(header >> 10) // Upper 6 bits

	useAscii := false
	if asciiIndicator == 0x3F {
		useAscii = true
		byteSize = (^byteSize) & 0x3FF // Invert byte size bits
	}

	if len(data) < 2 {
		return ""
	}

	textBytes := data[2:]
	hexChars := encoding.ByteArrayToHex(textBytes)

	if useAscii {
		return decodeHexAscii(hexChars, byteSize)
	}

	decoded, err := encoding.DecodeHexStarsString(hexChars, byteSize)
	if err != nil {
		return ""
	}
	return decoded
}

// decodeHexAscii decodes ASCII-encoded hex string
func decodeHexAscii(hexChars string, byteSize int) string {
	bytes := encoding.HexToByteArray(hexChars)
	if byteSize > len(bytes) {
		byteSize = len(bytes)
	}
	return string(bytes[:byteSize])
}

// Encode returns the raw block data bytes (without the 2-byte block header).
func (mb *MessageBlock) Encode() []byte {
	// Encode the message text
	messageEncoded := encodeStarsMessage(mb.Message)

	data := make([]byte, 10+len(messageEncoded))

	// Bytes 0-3: Write zeros (garbage bytes, ignored on read)
	encoding.Write16(data, 0, 0)
	encoding.Write16(data, 2, 0)
	encoding.Write16(data, 4, uint16(mb.SenderId))
	encoding.Write16(data, 6, uint16(mb.ReceiverId))
	encoding.Write16(data, 8, uint16(mb.InReplyTo))

	copy(data[10:], messageEncoded)

	return data
}

// encodeStarsMessage encodes a message string to Stars! format.
func encodeStarsMessage(message string) []byte {
	if message == "" {
		return []byte{0, 0} // Empty message
	}

	// Encode the string to hex chars (without length prefix)
	hexChars := encoding.EncodeHexStarsString(message)

	// Pad to even length if needed
	if len(hexChars)%2 != 0 {
		hexChars += "F"
	}

	// Convert hex chars to bytes
	textBytes := encoding.HexToByteArray(hexChars)

	// Build header: byteSize in lower 10 bits
	byteSize := len(textBytes)
	if byteSize > 0x3FF {
		byteSize = 0x3FF
	}

	header := uint16(byteSize)

	result := make([]byte, 2+len(textBytes))
	encoding.Write16(result, 0, header)
	copy(result[2:], textBytes)

	return result
}

// IsBroadcast returns true if the message was sent to everyone
func (mb *MessageBlock) IsBroadcast() bool {
	return mb.ReceiverId == 0
}

// IsReply returns true if this message is a reply to another message
func (mb *MessageBlock) IsReply() bool {
	return mb.InReplyTo > 0
}

// SenderDisplayId returns the 1-based player number of the sender
func (mb *MessageBlock) SenderDisplayId() int {
	return mb.SenderId + 1
}

// ReceiverDisplayId returns the 1-based player number of the receiver
// Returns 0 for broadcast messages
func (mb *MessageBlock) ReceiverDisplayId() int {
	return mb.ReceiverId
}
