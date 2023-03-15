use cosmwasm_std::{StdResult, Uint64, QuerierWrapper};

use crate::query::{BabylonQuery, CurrentEpochResponse};

pub struct BabylonQuerier<'a> {
  querier: &'a QuerierWrapper<'a, BabylonQuery>,
}

impl<'a> BabylonQuerier<'a> {
  pub fn new(querier: &'a QuerierWrapper<BabylonQuery>) -> Self {
    BabylonQuerier { querier }
  }

  pub fn current_epoch(&self) -> StdResult<Uint64> {
    let request = BabylonQuery::Epoch{}.into();
    let res: CurrentEpochResponse = self.querier.query(&request)?;
    return Ok(Uint64::new(res.epoch));
  }
}
