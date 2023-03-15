use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{CustomQuery};

#[cw_serde]
#[derive(QueryResponses)]
pub enum BabylonQuery {
  #[returns(CurrentEpochResponse)]
  Epoch {}
}


#[cw_serde]
pub struct CurrentEpochResponse {
    pub epoch: u64,
}

impl CustomQuery for BabylonQuery {}

impl BabylonQuery {
  pub fn current_epoch() -> Self {
    BabylonQuery::Epoch{}
  }
}
