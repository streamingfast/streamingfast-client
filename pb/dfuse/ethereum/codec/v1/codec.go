package pbcodec

import (
	"encoding/hex"

	"github.com/dfuse-io/bstream"
	"github.com/dfuse-io/jsonpb"
)

func (b *Block) ID() string {
	return hex.EncodeToString(b.Hash)
}

func (b *Block) PreviousID() string {
	return hex.EncodeToString(b.Header.ParentHash)
}

func (b *Block) AsRef() bstream.BlockRef {
	return bstream.NewBlockRef(b.ID(), b.Number)
}

func (m *BigInt) MarshalJSON() ([]byte, error) {
	if m == nil {
		// FIXME: What is the right behavior regarding JSON to output when there is no bytes? Usually I think it should be omitted
		//        entirely but I'm not sure what a custom JSON marshaler can do here to convey that meaning of ok, omit this field.
		return nil, nil
	}

	return []byte(`"` + hex.EncodeToString(m.Bytes) + `"`), nil
}

func (m *BigInt) MarshalJSONPB(marshaler *jsonpb.Marshaler) ([]byte, error) {
	return m.MarshalJSON()
}
