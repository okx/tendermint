package types

// state-delta mode
// 0 same as no state-delta
// 1 product delta and save into deltastore.db
// 2 consume delta and save into deltastore.db; if get no delta, do as 1
const (
	// for getting flag of delta-mode
	FlagStateDelta = "state-sync-mode"

	// delta-mode
	NoDelta      = "na"
	ProductDelta = "producer"
	ConsumeDelta = "consumer"

	// data-center
	FlagDataCenter = "data-center-mode"
	DataCenterUrl = "data-center-url"
	DataCenterStr = "dataCenter"
)

type BlockDelta struct {
	Block  *Block  `json:"block"`
	Deltas *Deltas `json:"deltas"`
	Height int64   `json:"height"`
}

// Deltas defines the ABCIResponse and state delta
type Deltas struct {
	ABCIRsp     []byte `json:"abci_rsp"`
	DeltasBytes []byte `json:"deltas_bytes"`
	Height      int64  `json:"height"`
}

// Size returns size of the deltas in bytes.
func (d *Deltas) Size() int {
	if d == nil {
		return 0
	}
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
