// @generated
impl serde::Serialize for BtcCheckpointInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.epoch_number != 0 {
            len += 1;
        }
        if self.earliest_btc_block_number != 0 {
            len += 1;
        }
        if !self.earliest_btc_block_hash.is_empty() {
            len += 1;
        }
        if !self.earliest_btc_block_txs.is_empty() {
            len += 1;
        }
        if !self.vigilante_address_list.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.btccheckpoint.v1.BTCCheckpointInfo", len)?;
        if self.epoch_number != 0 {
            struct_ser.serialize_field("epochNumber", ToString::to_string(&self.epoch_number).as_str())?;
        }
        if self.earliest_btc_block_number != 0 {
            struct_ser.serialize_field("earliestBtcBlockNumber", ToString::to_string(&self.earliest_btc_block_number).as_str())?;
        }
        if !self.earliest_btc_block_hash.is_empty() {
            struct_ser.serialize_field("earliestBtcBlockHash", pbjson::private::base64::encode(&self.earliest_btc_block_hash).as_str())?;
        }
        if !self.earliest_btc_block_txs.is_empty() {
            struct_ser.serialize_field("earliestBtcBlockTxs", &self.earliest_btc_block_txs)?;
        }
        if !self.vigilante_address_list.is_empty() {
            struct_ser.serialize_field("vigilanteAddressList", &self.vigilante_address_list)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for BtcCheckpointInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "epoch_number",
            "epochNumber",
            "earliest_btc_block_number",
            "earliestBtcBlockNumber",
            "earliest_btc_block_hash",
            "earliestBtcBlockHash",
            "earliest_btc_block_txs",
            "earliestBtcBlockTxs",
            "vigilante_address_list",
            "vigilanteAddressList",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EpochNumber,
            EarliestBtcBlockNumber,
            EarliestBtcBlockHash,
            EarliestBtcBlockTxs,
            VigilanteAddressList,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "epochNumber" | "epoch_number" => Ok(GeneratedField::EpochNumber),
                            "earliestBtcBlockNumber" | "earliest_btc_block_number" => Ok(GeneratedField::EarliestBtcBlockNumber),
                            "earliestBtcBlockHash" | "earliest_btc_block_hash" => Ok(GeneratedField::EarliestBtcBlockHash),
                            "earliestBtcBlockTxs" | "earliest_btc_block_txs" => Ok(GeneratedField::EarliestBtcBlockTxs),
                            "vigilanteAddressList" | "vigilante_address_list" => Ok(GeneratedField::VigilanteAddressList),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = BtcCheckpointInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.btccheckpoint.v1.BTCCheckpointInfo")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<BtcCheckpointInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut epoch_number__ = None;
                let mut earliest_btc_block_number__ = None;
                let mut earliest_btc_block_hash__ = None;
                let mut earliest_btc_block_txs__ = None;
                let mut vigilante_address_list__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::EpochNumber => {
                            if epoch_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochNumber"));
                            }
                            epoch_number__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EarliestBtcBlockNumber => {
                            if earliest_btc_block_number__.is_some() {
                                return Err(serde::de::Error::duplicate_field("earliestBtcBlockNumber"));
                            }
                            earliest_btc_block_number__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EarliestBtcBlockHash => {
                            if earliest_btc_block_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("earliestBtcBlockHash"));
                            }
                            earliest_btc_block_hash__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::EarliestBtcBlockTxs => {
                            if earliest_btc_block_txs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("earliestBtcBlockTxs"));
                            }
                            earliest_btc_block_txs__ = Some(map.next_value()?);
                        }
                        GeneratedField::VigilanteAddressList => {
                            if vigilante_address_list__.is_some() {
                                return Err(serde::de::Error::duplicate_field("vigilanteAddressList"));
                            }
                            vigilante_address_list__ = Some(map.next_value()?);
                        }
                    }
                }
                Ok(BtcCheckpointInfo {
                    epoch_number: epoch_number__.unwrap_or_default(),
                    earliest_btc_block_number: earliest_btc_block_number__.unwrap_or_default(),
                    earliest_btc_block_hash: earliest_btc_block_hash__.unwrap_or_default(),
                    earliest_btc_block_txs: earliest_btc_block_txs__.unwrap_or_default(),
                    vigilante_address_list: vigilante_address_list__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.btccheckpoint.v1.BTCCheckpointInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for BtcSpvProof {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.btc_transaction.is_empty() {
            len += 1;
        }
        if self.btc_transaction_index != 0 {
            len += 1;
        }
        if !self.merkle_nodes.is_empty() {
            len += 1;
        }
        if !self.confirming_btc_header.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.btccheckpoint.v1.BTCSpvProof", len)?;
        if !self.btc_transaction.is_empty() {
            struct_ser.serialize_field("btcTransaction", pbjson::private::base64::encode(&self.btc_transaction).as_str())?;
        }
        if self.btc_transaction_index != 0 {
            struct_ser.serialize_field("btcTransactionIndex", &self.btc_transaction_index)?;
        }
        if !self.merkle_nodes.is_empty() {
            struct_ser.serialize_field("merkleNodes", pbjson::private::base64::encode(&self.merkle_nodes).as_str())?;
        }
        if !self.confirming_btc_header.is_empty() {
            struct_ser.serialize_field("confirmingBtcHeader", pbjson::private::base64::encode(&self.confirming_btc_header).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for BtcSpvProof {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "btc_transaction",
            "btcTransaction",
            "btc_transaction_index",
            "btcTransactionIndex",
            "merkle_nodes",
            "merkleNodes",
            "confirming_btc_header",
            "confirmingBtcHeader",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            BtcTransaction,
            BtcTransactionIndex,
            MerkleNodes,
            ConfirmingBtcHeader,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "btcTransaction" | "btc_transaction" => Ok(GeneratedField::BtcTransaction),
                            "btcTransactionIndex" | "btc_transaction_index" => Ok(GeneratedField::BtcTransactionIndex),
                            "merkleNodes" | "merkle_nodes" => Ok(GeneratedField::MerkleNodes),
                            "confirmingBtcHeader" | "confirming_btc_header" => Ok(GeneratedField::ConfirmingBtcHeader),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = BtcSpvProof;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.btccheckpoint.v1.BTCSpvProof")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<BtcSpvProof, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut btc_transaction__ = None;
                let mut btc_transaction_index__ = None;
                let mut merkle_nodes__ = None;
                let mut confirming_btc_header__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::BtcTransaction => {
                            if btc_transaction__.is_some() {
                                return Err(serde::de::Error::duplicate_field("btcTransaction"));
                            }
                            btc_transaction__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BtcTransactionIndex => {
                            if btc_transaction_index__.is_some() {
                                return Err(serde::de::Error::duplicate_field("btcTransactionIndex"));
                            }
                            btc_transaction_index__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MerkleNodes => {
                            if merkle_nodes__.is_some() {
                                return Err(serde::de::Error::duplicate_field("merkleNodes"));
                            }
                            merkle_nodes__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::ConfirmingBtcHeader => {
                            if confirming_btc_header__.is_some() {
                                return Err(serde::de::Error::duplicate_field("confirmingBtcHeader"));
                            }
                            confirming_btc_header__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(BtcSpvProof {
                    btc_transaction: btc_transaction__.unwrap_or_default(),
                    btc_transaction_index: btc_transaction_index__.unwrap_or_default(),
                    merkle_nodes: merkle_nodes__.unwrap_or_default(),
                    confirming_btc_header: confirming_btc_header__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.btccheckpoint.v1.BTCSpvProof", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for BtcStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::EpochStatusSubmitted => "EPOCH_STATUS_SUBMITTED",
            Self::EpochStatusConfirmed => "EPOCH_STATUS_CONFIRMED",
            Self::EpochStatusFinalized => "EPOCH_STATUS_FINALIZED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for BtcStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "EPOCH_STATUS_SUBMITTED",
            "EPOCH_STATUS_CONFIRMED",
            "EPOCH_STATUS_FINALIZED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = BtcStatus;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                write!(formatter, "expected one of: {:?}", &FIELDS)
            }

            fn visit_i64<E>(self, v: i64) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                use std::convert::TryFrom;
                i32::try_from(v)
                    .ok()
                    .and_then(BtcStatus::from_i32)
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Signed(v), &self)
                    })
            }

            fn visit_u64<E>(self, v: u64) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                use std::convert::TryFrom;
                i32::try_from(v)
                    .ok()
                    .and_then(BtcStatus::from_i32)
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Unsigned(v), &self)
                    })
            }

            fn visit_str<E>(self, value: &str) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                match value {
                    "EPOCH_STATUS_SUBMITTED" => Ok(BtcStatus::EpochStatusSubmitted),
                    "EPOCH_STATUS_CONFIRMED" => Ok(BtcStatus::EpochStatusConfirmed),
                    "EPOCH_STATUS_FINALIZED" => Ok(BtcStatus::EpochStatusFinalized),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for CheckpointAddresses {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.submitter.is_empty() {
            len += 1;
        }
        if !self.reporter.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.btccheckpoint.v1.CheckpointAddresses", len)?;
        if !self.submitter.is_empty() {
            struct_ser.serialize_field("submitter", pbjson::private::base64::encode(&self.submitter).as_str())?;
        }
        if !self.reporter.is_empty() {
            struct_ser.serialize_field("reporter", pbjson::private::base64::encode(&self.reporter).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for CheckpointAddresses {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "submitter",
            "reporter",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Submitter,
            Reporter,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "submitter" => Ok(GeneratedField::Submitter),
                            "reporter" => Ok(GeneratedField::Reporter),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = CheckpointAddresses;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.btccheckpoint.v1.CheckpointAddresses")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<CheckpointAddresses, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut submitter__ = None;
                let mut reporter__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::Submitter => {
                            if submitter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("submitter"));
                            }
                            submitter__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Reporter => {
                            if reporter__.is_some() {
                                return Err(serde::de::Error::duplicate_field("reporter"));
                            }
                            reporter__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(CheckpointAddresses {
                    submitter: submitter__.unwrap_or_default(),
                    reporter: reporter__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.btccheckpoint.v1.CheckpointAddresses", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for EpochData {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.key.is_empty() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.btccheckpoint.v1.EpochData", len)?;
        if !self.key.is_empty() {
            struct_ser.serialize_field("key", &self.key)?;
        }
        if self.status != 0 {
            let v = BtcStatus::from_i32(self.status)
                .ok_or_else(|| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for EpochData {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "key",
            "status",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Key,
            Status,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "key" => Ok(GeneratedField::Key),
                            "status" => Ok(GeneratedField::Status),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = EpochData;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.btccheckpoint.v1.EpochData")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<EpochData, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut key__ = None;
                let mut status__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::Key => {
                            if key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("key"));
                            }
                            key__ = Some(map.next_value()?);
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map.next_value::<BtcStatus>()? as i32);
                        }
                    }
                }
                Ok(EpochData {
                    key: key__.unwrap_or_default(),
                    status: status__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.btccheckpoint.v1.EpochData", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for SubmissionData {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.vigilante_addresses.is_some() {
            len += 1;
        }
        if !self.txs_info.is_empty() {
            len += 1;
        }
        if self.epoch != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.btccheckpoint.v1.SubmissionData", len)?;
        if let Some(v) = self.vigilante_addresses.as_ref() {
            struct_ser.serialize_field("vigilanteAddresses", v)?;
        }
        if !self.txs_info.is_empty() {
            struct_ser.serialize_field("txsInfo", &self.txs_info)?;
        }
        if self.epoch != 0 {
            struct_ser.serialize_field("epoch", ToString::to_string(&self.epoch).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for SubmissionData {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "vigilante_addresses",
            "vigilanteAddresses",
            "txs_info",
            "txsInfo",
            "epoch",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            VigilanteAddresses,
            TxsInfo,
            Epoch,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "vigilanteAddresses" | "vigilante_addresses" => Ok(GeneratedField::VigilanteAddresses),
                            "txsInfo" | "txs_info" => Ok(GeneratedField::TxsInfo),
                            "epoch" => Ok(GeneratedField::Epoch),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = SubmissionData;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.btccheckpoint.v1.SubmissionData")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<SubmissionData, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut vigilante_addresses__ = None;
                let mut txs_info__ = None;
                let mut epoch__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::VigilanteAddresses => {
                            if vigilante_addresses__.is_some() {
                                return Err(serde::de::Error::duplicate_field("vigilanteAddresses"));
                            }
                            vigilante_addresses__ = map.next_value()?;
                        }
                        GeneratedField::TxsInfo => {
                            if txs_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("txsInfo"));
                            }
                            txs_info__ = Some(map.next_value()?);
                        }
                        GeneratedField::Epoch => {
                            if epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epoch"));
                            }
                            epoch__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(SubmissionData {
                    vigilante_addresses: vigilante_addresses__,
                    txs_info: txs_info__.unwrap_or_default(),
                    epoch: epoch__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.btccheckpoint.v1.SubmissionData", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for SubmissionKey {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.key.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.btccheckpoint.v1.SubmissionKey", len)?;
        if !self.key.is_empty() {
            struct_ser.serialize_field("key", &self.key)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for SubmissionKey {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "key",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Key,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "key" => Ok(GeneratedField::Key),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = SubmissionKey;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.btccheckpoint.v1.SubmissionKey")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<SubmissionKey, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut key__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::Key => {
                            if key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("key"));
                            }
                            key__ = Some(map.next_value()?);
                        }
                    }
                }
                Ok(SubmissionKey {
                    key: key__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.btccheckpoint.v1.SubmissionKey", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for TransactionInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.key.is_some() {
            len += 1;
        }
        if !self.transaction.is_empty() {
            len += 1;
        }
        if !self.proof.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.btccheckpoint.v1.TransactionInfo", len)?;
        if let Some(v) = self.key.as_ref() {
            struct_ser.serialize_field("key", v)?;
        }
        if !self.transaction.is_empty() {
            struct_ser.serialize_field("transaction", pbjson::private::base64::encode(&self.transaction).as_str())?;
        }
        if !self.proof.is_empty() {
            struct_ser.serialize_field("proof", pbjson::private::base64::encode(&self.proof).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for TransactionInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "key",
            "transaction",
            "proof",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Key,
            Transaction,
            Proof,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "key" => Ok(GeneratedField::Key),
                            "transaction" => Ok(GeneratedField::Transaction),
                            "proof" => Ok(GeneratedField::Proof),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = TransactionInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.btccheckpoint.v1.TransactionInfo")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<TransactionInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut key__ = None;
                let mut transaction__ = None;
                let mut proof__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::Key => {
                            if key__.is_some() {
                                return Err(serde::de::Error::duplicate_field("key"));
                            }
                            key__ = map.next_value()?;
                        }
                        GeneratedField::Transaction => {
                            if transaction__.is_some() {
                                return Err(serde::de::Error::duplicate_field("transaction"));
                            }
                            transaction__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Proof => {
                            if proof__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proof"));
                            }
                            proof__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(TransactionInfo {
                    key: key__,
                    transaction: transaction__.unwrap_or_default(),
                    proof: proof__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.btccheckpoint.v1.TransactionInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for TransactionKey {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.index != 0 {
            len += 1;
        }
        if !self.hash.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.btccheckpoint.v1.TransactionKey", len)?;
        if self.index != 0 {
            struct_ser.serialize_field("index", &self.index)?;
        }
        if !self.hash.is_empty() {
            struct_ser.serialize_field("hash", pbjson::private::base64::encode(&self.hash).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for TransactionKey {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "index",
            "hash",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Index,
            Hash,
        }
        impl<'de> serde::Deserialize<'de> for GeneratedField {
            fn deserialize<D>(deserializer: D) -> std::result::Result<GeneratedField, D::Error>
            where
                D: serde::Deserializer<'de>,
            {
                struct GeneratedVisitor;

                impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
                    type Value = GeneratedField;

                    fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                        write!(formatter, "expected one of: {:?}", &FIELDS)
                    }

                    #[allow(unused_variables)]
                    fn visit_str<E>(self, value: &str) -> std::result::Result<GeneratedField, E>
                    where
                        E: serde::de::Error,
                    {
                        match value {
                            "index" => Ok(GeneratedField::Index),
                            "hash" => Ok(GeneratedField::Hash),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = TransactionKey;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.btccheckpoint.v1.TransactionKey")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<TransactionKey, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut index__ = None;
                let mut hash__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::Index => {
                            if index__.is_some() {
                                return Err(serde::de::Error::duplicate_field("index"));
                            }
                            index__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Hash => {
                            if hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hash"));
                            }
                            hash__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(TransactionKey {
                    index: index__.unwrap_or_default(),
                    hash: hash__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.btccheckpoint.v1.TransactionKey", FIELDS, GeneratedVisitor)
    }
}
