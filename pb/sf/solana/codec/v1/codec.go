package pbcodec

import (
	"encoding/hex"

	sf "github.com/streamingfast/streamingfast-client"
)

func (b *Block) ID() string {
	return hex.EncodeToString(b.Id)
}

func (b *Block) PreviousID() string {
	return hex.EncodeToString(b.PreviousId)
}

func (b *Block) AsRef() sf.BlockRef {
	return sf.NewBlockRef(b.ID(), b.Number)
}
