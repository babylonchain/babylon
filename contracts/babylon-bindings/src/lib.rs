mod query;
mod querier;

pub use query::{BabylonQuery, CurrentEpochResponse};
pub use querier::BabylonQuerier;
// This export is added to all contracts that import this package, signifying that they require
// "terra" support on the chain they run on.
#[no_mangle]
extern "C" fn requires_babylon() {}
