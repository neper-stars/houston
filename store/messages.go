package store

import "github.com/neper-stars/houston/blocks"

// MessageEntity represents a player message.
type MessageEntity struct {
	meta EntityMeta

	// Message data
	SenderId   int    // Sender player number (0-15)
	ReceiverId int    // Receiver: 0=everyone, 1-16=specific player
	Message    string // Message text

	// Raw block (preserved for re-encoding)
	messageBlock *blocks.MessageBlock
}

// Meta returns the entity metadata.
func (m *MessageEntity) Meta() *EntityMeta {
	return &m.meta
}

// RawBlocks returns the original blocks.
func (m *MessageEntity) RawBlocks() []blocks.Block {
	if m.messageBlock != nil {
		return []blocks.Block{*m.messageBlock}
	}
	return nil
}

// SetDirty marks the entity as modified.
func (m *MessageEntity) SetDirty() {
	m.meta.Dirty = true
}

// IsBroadcast returns true if the message is sent to everyone.
func (m *MessageEntity) IsBroadcast() bool {
	return m.ReceiverId == 0
}

// newMessageEntityFromBlock creates a MessageEntity from a MessageBlock.
func newMessageEntityFromBlock(mb *blocks.MessageBlock, index int, source *FileSource) *MessageEntity {
	entity := &MessageEntity{
		meta: EntityMeta{
			Key: EntityKey{
				Type:   EntityTypeMessage,
				Owner:  mb.SenderId,
				Number: index, // Use index as message number
			},
			BestSource: source,
			Quality:    QualityFull,
			Turn:       source.Turn,
		},
		SenderId:     mb.SenderId,
		ReceiverId:   mb.ReceiverId,
		Message:      mb.Message,
		messageBlock: mb,
	}
	entity.meta.AddSource(source)
	return entity
}
