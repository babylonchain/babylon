package bindings

type BabylonQuery struct {
	Epoch *struct{} `json:"epoch,omitempty"`
}

type CurrentEpochResponse struct {
	Epoch uint64 `json:"epoch"`
}

// type StakingQuery struct {
// 	AllValidators  *AllValidatorsQuery  `json:"all_validators,omitempty"`
// 	Validator      *ValidatorQuery      `json:"validator,omitempty"`
// 	AllDelegations *AllDelegationsQuery `json:"all_delegations,omitempty"`
// 	Delegation     *DelegationQuery     `json:"delegation,omitempty"`
// 	BondedDenom    *struct{}            `json:"bonded_denom,omitempty"`
// }
