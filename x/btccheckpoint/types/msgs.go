package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	// Ensure that MsgInsertBTCSpvProof implements all functions of the Msg interface
	_ sdk.Msg = (*MsgInsertBTCSpvProof)(nil)
)

func (m *MsgInsertBTCSpvProof) ValidateBasic() error {
	// TODO: Implement me
	return nil
}

func (m *MsgInsertBTCSpvProof) GetSigners() []sdk.AccAddress {
	// TODO: implement me
	return []sdk.AccAddress{}
}
