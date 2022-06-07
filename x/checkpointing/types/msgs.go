package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	// Ensure that MsgInsertHeader implements all functions of the Msg interface
	_ sdk.Msg = (*MsgCollectBlsSig)(nil)
)

func (m *MsgCollectBlsSig) ValidateBasic() error {
	// This function validates stateless message elements
	_, err := sdk.AccAddressFromBech32(m.BlsSig.Signer)
	if err != nil {
		return err
	}

	// TODO: verify bls sig

	return nil
}

func (m *MsgCollectBlsSig) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.BlsSig.Signer)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{signer}
}
