package pbcodec

import (
	"encoding/hex"

	"github.com/streamingfast/bstream"
)

func (b *Block) ID() string {
	return hex.EncodeToString(b.Id)
}

func (b *Block) PreviousID() string {
	return hex.EncodeToString(b.PreviousId)
}

func (b *Block) AsRef() bstream.BlockRef {
	return bstream.NewBlockRef(b.ID(), b.Number)
}
