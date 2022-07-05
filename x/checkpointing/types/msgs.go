package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	// Ensure that MsgInsertHeader implements all functions of the Msg interface
	_ sdk.Msg = (*MsgAddBlsSig)(nil)
)

func (m *MsgAddBlsSig) ValidateBasic() error {
	// This function validates stateless message elements
	_, err := sdk.AccAddressFromBech32(m.BlsSig.SignerAddress)
	if err != nil {
		return err
	}

	// TODO: verify bls sig

	return nil
}

func (m *MsgAddBlsSig) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.BlsSig.SignerAddress)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{signer}
}

func (m *MsgWrappedCreateValidator) ValidateBasic() error {
	// This function validates stateless message elements
	_, err := sdk.AccAddressFromBech32(m.Pubkey.Address)
	if err != nil {
		return err
	}

	// TODO: verify bls sig

	return nil
}

func (m *MsgWrappedCreateValidator) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(m.Pubkey.Address)
	if err != nil {
		panic(err)
	}

	return []sdk.AccAddress{signer}
}
