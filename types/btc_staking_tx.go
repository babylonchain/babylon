package types

type BTCStakingTx []byte

// TODO: constructors, verification rules, and util functions

func (tx BTCStakingTx) Marshal() ([]byte, error) {
	return tx, nil
}

func (tx BTCStakingTx) MustMarshal() []byte {
	txBytes, err := tx.Marshal()
	if err != nil {
		panic(err)
	}
	return txBytes
}

func (tx BTCStakingTx) MarshalTo(data []byte) (int, error) {
	bz, err := tx.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(data), nil
}

func (tx *BTCStakingTx) Unmarshal(data []byte) error {
	// TODO: verifications

	*tx = data
	return nil
}

func (tx *BTCStakingTx) Size() int {
	return len(tx.MustMarshal())
}
