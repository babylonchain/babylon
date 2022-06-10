package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
    // Ensure that MsgInsertHeader implements all functions of the Msg interface
	_ sdk.Msg = (*MsgInsertHeader)(nil)
)

func (m *MsgInsertHeader) ValidateBasic() error {
    // This function validates stateless message elements
	_, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
		return err
	}

	if m.Header == nil {
		return ErrInvalidHeader.Wrap("empty")
	}

	if len(m.Header.ParentHash) != 32 {
		return ErrInvalidHeader.Wrap("parent hash size not 32")
	}

	if len(m.Header.MerkleRoot) != 32 {
		return ErrInvalidHeader.Wrap("merkle root size not 32")
	}

	// TODO: verify version proper value

	// TODO: verify difficulty

	return nil
}

func (m *MsgInsertHeader) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Signer)
	if err != nil {
        // Panic, since the GetSigners method is called after ValidateBasic
        // which performs the same check.
		panic(err)
	}

	return []sdk.AccAddress{signer}
}
