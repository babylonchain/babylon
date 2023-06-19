package types

type BTCSlashingTx []byte

// TODO: constructors, verification rules, and util functions

func (tx BTCSlashingTx) Marshal() ([]byte, error) {
	return tx, nil
}

func (tx BTCSlashingTx) MustMarshal() []byte {
	txBytes, err := tx.Marshal()
	if err != nil {
		panic(err)
	}
	return txBytes
}

func (tx BTCSlashingTx) MarshalTo(data []byte) (int, error) {
	bz, err := tx.Marshal()
	if err != nil {
		return 0, err
	}
	copy(data, bz)
	return len(data), nil
}

func (tx *BTCSlashingTx) Unmarshal(data []byte) error {
	// TODO: verifications

	*tx = data
	return nil
}

func (tx *BTCSlashingTx) Size() int {
	return len(tx.MustMarshal())
}
