use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Uint64;

/// Just needs to know the code_id of a reflect contract to spawn sub-accounts
#[cw_serde]
pub struct InstantiateMsg {
    pub reflect_code_id: u64,
}

#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    #[returns(GetCurrentEpochResponse)]
    GetCurrentEpoch
}

#[cw_serde]
pub struct GetCurrentEpochResponse{
    pub epoch: Uint64
}
