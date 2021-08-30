package types

// Deltas defines the ABCIResponse and state delta
type Deltas struct {
	ABCIRsp	[]byte
	DeltasBytes	[]byte
	Height  int64
}

// Size returns size of the deltas in bytes.
func (d *Deltas) Size() int {
	bz, err := cdc.MarshalBinaryBare(d)
	if err != nil {
		return 0
	}
	return len(bz)
}

// Marshal returns the amino encoding.
func (b *Deltas) Marshal() ([]byte, error) {
	return cdc.MarshalBinaryBare(b)
}

// Unmarshal deserializes from amino encoded form.
func (b *Deltas) Unmarshal(bs []byte) error {
	return cdc.UnmarshalBinaryBare(bs, b)
}
