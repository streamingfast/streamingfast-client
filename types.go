package sf

import "fmt"

var EmptyBlockRef = BlockRef{ID: "", Number: 0}

// blockRef is just a thin wrapper to carry over the block number and its id
// as well as providing a simple enough `String` implementation for logging
// or display purposes.
//
// We do not use `bstream.BlockRef` mainly to avoid pulling `bstream` library
// which pulls `dstore` which pulls `github.com/Azure/azure-pipeline-go` which
// pulls `github.com/mattn/go-ieproxy` which creates cross-compilation problems.
type BlockRef struct {
	ID     string
	Number uint64
}

func NewBlockRef(id string, number uint64) BlockRef {
	return BlockRef{ID: id, Number: number}
}

func (b BlockRef) String() string {
	return fmt.Sprintf("#%d (%s)", b.Number, b.ID)
}
