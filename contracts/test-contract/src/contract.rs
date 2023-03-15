use std::str::FromStr;

use cosmwasm_std::{
    entry_point, to_binary, Deps, DepsMut, Env, MessageInfo, QueryResponse, Response, StdResult, CustomQuery,
};

use crate::msg::{
    InstantiateMsg, QueryMsg, GetCurrentEpochResponse
};
use crate::state::{config, Config};

use babylon_bindings::{BabylonQuerier, BabylonQuery, CurrentEpochResponse};

pub const IBC_APP_VERSION: &str = "babylon-contract-v1";
pub const RECEIVE_DISPATCH_ID: u64 = 1234;
pub const INIT_CALLBACK_ID: u64 = 7890;

#[entry_point]
pub fn instantiate(
    deps: DepsMut<BabylonQuery>,
    _env: Env,
    _info: MessageInfo,
    msg: InstantiateMsg,
) -> StdResult<Response> {
    // we store the reflect_id for creating accounts later
    let cfg = Config {
        placeholder_id: msg.reflect_code_id,
    };
    config(deps.storage).save(&cfg)?;

    Ok(Response::new().add_attribute("action", "instantiate"))
}

#[entry_point]
pub fn query(deps: Deps<BabylonQuery>, _env: Env, msg: QueryMsg) -> StdResult<QueryResponse> {
    match msg {
        QueryMsg::GetCurrentEpoch => to_binary(&get_epoch(deps)?),
    }
}

pub fn get_epoch(deps: Deps<BabylonQuery>) -> StdResult<GetCurrentEpochResponse> {
    let bquerier = BabylonQuerier::new(&deps.querier);
    let resp = bquerier.current_epoch()?;
    return Ok(GetCurrentEpochResponse{epoch: resp});
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{
        mock_env, mock_info, MockApi, MockQuerier,
        MockStorage,
    };
    use cosmwasm_std::{OwnedDeps, Uint64};
    use std::marker::PhantomData;

    use cosmwasm_std::testing::{MOCK_CONTRACT_ADDR};

    use cosmwasm_std::{to_binary, Binary, ContractResult, SystemResult};

    const CREATOR: &str = "creator";
    // code id of the reflect contract
    const REFLECT_ID: u64 = 101;

    pub fn mock_dependencies_with_custom_querier(
    ) -> OwnedDeps<MockStorage, MockApi, MockQuerier<BabylonQuery>, BabylonQuery> {
        let custom_querier: MockQuerier<BabylonQuery> =
            MockQuerier::new(&[(MOCK_CONTRACT_ADDR, &[])])
                .with_custom_handler(|query| SystemResult::Ok(custom_query_execute(query)));
        OwnedDeps {
            storage: MockStorage::default(),
            api: MockApi::default(),
            querier: custom_querier,
            custom_query_type: PhantomData,
        }
    }

    pub fn custom_query_execute(query: &BabylonQuery) -> ContractResult<Binary> {
        let msg: Uint64 = match query {
            BabylonQuery::Epoch {} => Uint64::from(43u64),
        };

        let resp = CurrentEpochResponse {epoch: msg};
        to_binary(&resp).into()
    }


    #[test]
    fn instantiate_works() {
        let mut deps = mock_dependencies_with_custom_querier();

        let msg = InstantiateMsg {
            reflect_code_id: 17,
        };
        let info = mock_info("creator", &[]);
        let res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(0, res.messages.len())
    }
}
