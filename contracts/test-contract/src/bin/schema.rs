use std::env::current_dir;

use cosmwasm_schema::{export_schema, export_schema_with_title, schema_for, write_api};
use cosmwasm_std::Empty;

use babylon_contract::msg::{
    InstantiateMsg, QueryMsg
};

fn main() {
    // Clear & write standard API
    write_api! {
        instantiate: InstantiateMsg,
        query: QueryMsg,
        migrate: Empty,
    }

    // // Schemas for inter-contract communication
    // let mut out_dir = current_dir().unwrap();
    // out_dir.push("schema");
    // export_schema(&schema_for!(PacketMsg), &out_dir);
    // export_schema_with_title(
    //     &schema_for!(AcknowledgementMsg<WhoAmIResponse>),
    //     &out_dir,
    //     "AcknowledgementMsgWhoAmI",
    // );
}
