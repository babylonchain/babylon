// @generated
impl serde::Serialize for BondState {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::Created => "CREATED",
            Self::Bonded => "BONDED",
            Self::Unbonding => "UNBONDING",
            Self::Unbonded => "UNBONDED",
            Self::Removed => "REMOVED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for BondState {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "CREATED",
            "BONDED",
            "UNBONDING",
            "UNBONDED",
            "REMOVED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = BondState;

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
                    .and_then(BondState::from_i32)
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
                    .and_then(BondState::from_i32)
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Unsigned(v), &self)
                    })
            }

            fn visit_str<E>(self, value: &str) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                match value {
                    "CREATED" => Ok(BondState::Created),
                    "BONDED" => Ok(BondState::Bonded),
                    "UNBONDING" => Ok(BondState::Unbonding),
                    "UNBONDED" => Ok(BondState::Unbonded),
                    "REMOVED" => Ok(BondState::Removed),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for DelegationLifecycle {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.del_addr.is_empty() {
            len += 1;
        }
        if !self.del_life.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.epoching.v1.DelegationLifecycle", len)?;
        if !self.del_addr.is_empty() {
            struct_ser.serialize_field("delAddr", &self.del_addr)?;
        }
        if !self.del_life.is_empty() {
            struct_ser.serialize_field("delLife", &self.del_life)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for DelegationLifecycle {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "del_addr",
            "delAddr",
            "del_life",
            "delLife",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            DelAddr,
            DelLife,
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
                            "delAddr" | "del_addr" => Ok(GeneratedField::DelAddr),
                            "delLife" | "del_life" => Ok(GeneratedField::DelLife),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DelegationLifecycle;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.epoching.v1.DelegationLifecycle")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<DelegationLifecycle, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut del_addr__ = None;
                let mut del_life__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::DelAddr => {
                            if del_addr__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delAddr"));
                            }
                            del_addr__ = Some(map.next_value()?);
                        }
                        GeneratedField::DelLife => {
                            if del_life__.is_some() {
                                return Err(serde::de::Error::duplicate_field("delLife"));
                            }
                            del_life__ = Some(map.next_value()?);
                        }
                    }
                }
                Ok(DelegationLifecycle {
                    del_addr: del_addr__.unwrap_or_default(),
                    del_life: del_life__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.epoching.v1.DelegationLifecycle", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for DelegationStateUpdate {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.state != 0 {
            len += 1;
        }
        if !self.val_addr.is_empty() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        if self.block_time.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.epoching.v1.DelegationStateUpdate", len)?;
        if self.state != 0 {
            let v = BondState::from_i32(self.state)
                .ok_or_else(|| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if !self.val_addr.is_empty() {
            struct_ser.serialize_field("valAddr", &self.val_addr)?;
        }
        if self.block_height != 0 {
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        if let Some(v) = self.block_time.as_ref() {
            struct_ser.serialize_field("blockTime", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for DelegationStateUpdate {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "state",
            "val_addr",
            "valAddr",
            "block_height",
            "blockHeight",
            "block_time",
            "blockTime",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            State,
            ValAddr,
            BlockHeight,
            BlockTime,
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
                            "state" => Ok(GeneratedField::State),
                            "valAddr" | "val_addr" => Ok(GeneratedField::ValAddr),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            "blockTime" | "block_time" => Ok(GeneratedField::BlockTime),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = DelegationStateUpdate;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.epoching.v1.DelegationStateUpdate")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<DelegationStateUpdate, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut state__ = None;
                let mut val_addr__ = None;
                let mut block_height__ = None;
                let mut block_time__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map.next_value::<BondState>()? as i32);
                        }
                        GeneratedField::ValAddr => {
                            if val_addr__.is_some() {
                                return Err(serde::de::Error::duplicate_field("valAddr"));
                            }
                            val_addr__ = Some(map.next_value()?);
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlockTime => {
                            if block_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockTime"));
                            }
                            block_time__ = map.next_value()?;
                        }
                    }
                }
                Ok(DelegationStateUpdate {
                    state: state__.unwrap_or_default(),
                    val_addr: val_addr__.unwrap_or_default(),
                    block_height: block_height__.unwrap_or_default(),
                    block_time: block_time__,
                })
            }
        }
        deserializer.deserialize_struct("babylon.epoching.v1.DelegationStateUpdate", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Epoch {
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
        if self.current_epoch_interval != 0 {
            len += 1;
        }
        if self.first_block_height != 0 {
            len += 1;
        }
        if self.last_block_header.is_some() {
            len += 1;
        }
        if !self.app_hash_root.is_empty() {
            len += 1;
        }
        if self.sealer_header.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.epoching.v1.Epoch", len)?;
        if self.epoch_number != 0 {
            struct_ser.serialize_field("epochNumber", ToString::to_string(&self.epoch_number).as_str())?;
        }
        if self.current_epoch_interval != 0 {
            struct_ser.serialize_field("currentEpochInterval", ToString::to_string(&self.current_epoch_interval).as_str())?;
        }
        if self.first_block_height != 0 {
            struct_ser.serialize_field("firstBlockHeight", ToString::to_string(&self.first_block_height).as_str())?;
        }
        if let Some(v) = self.last_block_header.as_ref() {
            struct_ser.serialize_field("lastBlockHeader", v)?;
        }
        if !self.app_hash_root.is_empty() {
            struct_ser.serialize_field("appHashRoot", pbjson::private::base64::encode(&self.app_hash_root).as_str())?;
        }
        if let Some(v) = self.sealer_header.as_ref() {
            struct_ser.serialize_field("sealerHeader", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Epoch {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "epoch_number",
            "epochNumber",
            "current_epoch_interval",
            "currentEpochInterval",
            "first_block_height",
            "firstBlockHeight",
            "last_block_header",
            "lastBlockHeader",
            "app_hash_root",
            "appHashRoot",
            "sealer_header",
            "sealerHeader",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EpochNumber,
            CurrentEpochInterval,
            FirstBlockHeight,
            LastBlockHeader,
            AppHashRoot,
            SealerHeader,
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
                            "currentEpochInterval" | "current_epoch_interval" => Ok(GeneratedField::CurrentEpochInterval),
                            "firstBlockHeight" | "first_block_height" => Ok(GeneratedField::FirstBlockHeight),
                            "lastBlockHeader" | "last_block_header" => Ok(GeneratedField::LastBlockHeader),
                            "appHashRoot" | "app_hash_root" => Ok(GeneratedField::AppHashRoot),
                            "sealerHeader" | "sealer_header" => Ok(GeneratedField::SealerHeader),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Epoch;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.epoching.v1.Epoch")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<Epoch, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut epoch_number__ = None;
                let mut current_epoch_interval__ = None;
                let mut first_block_height__ = None;
                let mut last_block_header__ = None;
                let mut app_hash_root__ = None;
                let mut sealer_header__ = None;
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
                        GeneratedField::CurrentEpochInterval => {
                            if current_epoch_interval__.is_some() {
                                return Err(serde::de::Error::duplicate_field("currentEpochInterval"));
                            }
                            current_epoch_interval__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::FirstBlockHeight => {
                            if first_block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("firstBlockHeight"));
                            }
                            first_block_height__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastBlockHeader => {
                            if last_block_header__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastBlockHeader"));
                            }
                            last_block_header__ = map.next_value()?;
                        }
                        GeneratedField::AppHashRoot => {
                            if app_hash_root__.is_some() {
                                return Err(serde::de::Error::duplicate_field("appHashRoot"));
                            }
                            app_hash_root__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SealerHeader => {
                            if sealer_header__.is_some() {
                                return Err(serde::de::Error::duplicate_field("sealerHeader"));
                            }
                            sealer_header__ = map.next_value()?;
                        }
                    }
                }
                Ok(Epoch {
                    epoch_number: epoch_number__.unwrap_or_default(),
                    current_epoch_interval: current_epoch_interval__.unwrap_or_default(),
                    first_block_height: first_block_height__.unwrap_or_default(),
                    last_block_header: last_block_header__,
                    app_hash_root: app_hash_root__.unwrap_or_default(),
                    sealer_header: sealer_header__,
                })
            }
        }
        deserializer.deserialize_struct("babylon.epoching.v1.Epoch", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueuedMessage {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.tx_id.is_empty() {
            len += 1;
        }
        if !self.msg_id.is_empty() {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        if self.block_time.is_some() {
            len += 1;
        }
        if self.msg.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.epoching.v1.QueuedMessage", len)?;
        if !self.tx_id.is_empty() {
            struct_ser.serialize_field("txId", pbjson::private::base64::encode(&self.tx_id).as_str())?;
        }
        if !self.msg_id.is_empty() {
            struct_ser.serialize_field("msgId", pbjson::private::base64::encode(&self.msg_id).as_str())?;
        }
        if self.block_height != 0 {
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        if let Some(v) = self.block_time.as_ref() {
            struct_ser.serialize_field("blockTime", v)?;
        }
        if let Some(v) = self.msg.as_ref() {
            match v {
                queued_message::Msg::MsgCreateValidator(v) => {
                    struct_ser.serialize_field("msgCreateValidator", v)?;
                }
                queued_message::Msg::MsgDelegate(v) => {
                    struct_ser.serialize_field("msgDelegate", v)?;
                }
                queued_message::Msg::MsgUndelegate(v) => {
                    struct_ser.serialize_field("msgUndelegate", v)?;
                }
                queued_message::Msg::MsgBeginRedelegate(v) => {
                    struct_ser.serialize_field("msgBeginRedelegate", v)?;
                }
            }
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueuedMessage {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "tx_id",
            "txId",
            "msg_id",
            "msgId",
            "block_height",
            "blockHeight",
            "block_time",
            "blockTime",
            "msg_create_validator",
            "msgCreateValidator",
            "msg_delegate",
            "msgDelegate",
            "msg_undelegate",
            "msgUndelegate",
            "msg_begin_redelegate",
            "msgBeginRedelegate",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            TxId,
            MsgId,
            BlockHeight,
            BlockTime,
            MsgCreateValidator,
            MsgDelegate,
            MsgUndelegate,
            MsgBeginRedelegate,
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
                            "txId" | "tx_id" => Ok(GeneratedField::TxId),
                            "msgId" | "msg_id" => Ok(GeneratedField::MsgId),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            "blockTime" | "block_time" => Ok(GeneratedField::BlockTime),
                            "msgCreateValidator" | "msg_create_validator" => Ok(GeneratedField::MsgCreateValidator),
                            "msgDelegate" | "msg_delegate" => Ok(GeneratedField::MsgDelegate),
                            "msgUndelegate" | "msg_undelegate" => Ok(GeneratedField::MsgUndelegate),
                            "msgBeginRedelegate" | "msg_begin_redelegate" => Ok(GeneratedField::MsgBeginRedelegate),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueuedMessage;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.epoching.v1.QueuedMessage")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<QueuedMessage, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut tx_id__ = None;
                let mut msg_id__ = None;
                let mut block_height__ = None;
                let mut block_time__ = None;
                let mut msg__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::TxId => {
                            if tx_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("txId"));
                            }
                            tx_id__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::MsgId => {
                            if msg_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("msgId"));
                            }
                            msg_id__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlockTime => {
                            if block_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockTime"));
                            }
                            block_time__ = map.next_value()?;
                        }
                        GeneratedField::MsgCreateValidator => {
                            if msg__.is_some() {
                                return Err(serde::de::Error::duplicate_field("msgCreateValidator"));
                            }
                            msg__ = map.next_value::<::std::option::Option<_>>()?.map(queued_message::Msg::MsgCreateValidator)
;
                        }
                        GeneratedField::MsgDelegate => {
                            if msg__.is_some() {
                                return Err(serde::de::Error::duplicate_field("msgDelegate"));
                            }
                            msg__ = map.next_value::<::std::option::Option<_>>()?.map(queued_message::Msg::MsgDelegate)
;
                        }
                        GeneratedField::MsgUndelegate => {
                            if msg__.is_some() {
                                return Err(serde::de::Error::duplicate_field("msgUndelegate"));
                            }
                            msg__ = map.next_value::<::std::option::Option<_>>()?.map(queued_message::Msg::MsgUndelegate)
;
                        }
                        GeneratedField::MsgBeginRedelegate => {
                            if msg__.is_some() {
                                return Err(serde::de::Error::duplicate_field("msgBeginRedelegate"));
                            }
                            msg__ = map.next_value::<::std::option::Option<_>>()?.map(queued_message::Msg::MsgBeginRedelegate)
;
                        }
                    }
                }
                Ok(QueuedMessage {
                    tx_id: tx_id__.unwrap_or_default(),
                    msg_id: msg_id__.unwrap_or_default(),
                    block_height: block_height__.unwrap_or_default(),
                    block_time: block_time__,
                    msg: msg__,
                })
            }
        }
        deserializer.deserialize_struct("babylon.epoching.v1.QueuedMessage", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for QueuedMessageList {
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
        if !self.msgs.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.epoching.v1.QueuedMessageList", len)?;
        if self.epoch_number != 0 {
            struct_ser.serialize_field("epochNumber", ToString::to_string(&self.epoch_number).as_str())?;
        }
        if !self.msgs.is_empty() {
            struct_ser.serialize_field("msgs", &self.msgs)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for QueuedMessageList {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "epoch_number",
            "epochNumber",
            "msgs",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EpochNumber,
            Msgs,
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
                            "msgs" => Ok(GeneratedField::Msgs),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = QueuedMessageList;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.epoching.v1.QueuedMessageList")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<QueuedMessageList, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut epoch_number__ = None;
                let mut msgs__ = None;
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
                        GeneratedField::Msgs => {
                            if msgs__.is_some() {
                                return Err(serde::de::Error::duplicate_field("msgs"));
                            }
                            msgs__ = Some(map.next_value()?);
                        }
                    }
                }
                Ok(QueuedMessageList {
                    epoch_number: epoch_number__.unwrap_or_default(),
                    msgs: msgs__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.epoching.v1.QueuedMessageList", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ValStateUpdate {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.state != 0 {
            len += 1;
        }
        if self.block_height != 0 {
            len += 1;
        }
        if self.block_time.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.epoching.v1.ValStateUpdate", len)?;
        if self.state != 0 {
            let v = BondState::from_i32(self.state)
                .ok_or_else(|| serde::ser::Error::custom(format!("Invalid variant {}", self.state)))?;
            struct_ser.serialize_field("state", &v)?;
        }
        if self.block_height != 0 {
            struct_ser.serialize_field("blockHeight", ToString::to_string(&self.block_height).as_str())?;
        }
        if let Some(v) = self.block_time.as_ref() {
            struct_ser.serialize_field("blockTime", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ValStateUpdate {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "state",
            "block_height",
            "blockHeight",
            "block_time",
            "blockTime",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            State,
            BlockHeight,
            BlockTime,
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
                            "state" => Ok(GeneratedField::State),
                            "blockHeight" | "block_height" => Ok(GeneratedField::BlockHeight),
                            "blockTime" | "block_time" => Ok(GeneratedField::BlockTime),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ValStateUpdate;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.epoching.v1.ValStateUpdate")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<ValStateUpdate, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut state__ = None;
                let mut block_height__ = None;
                let mut block_time__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::State => {
                            if state__.is_some() {
                                return Err(serde::de::Error::duplicate_field("state"));
                            }
                            state__ = Some(map.next_value::<BondState>()? as i32);
                        }
                        GeneratedField::BlockHeight => {
                            if block_height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockHeight"));
                            }
                            block_height__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlockTime => {
                            if block_time__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blockTime"));
                            }
                            block_time__ = map.next_value()?;
                        }
                    }
                }
                Ok(ValStateUpdate {
                    state: state__.unwrap_or_default(),
                    block_height: block_height__.unwrap_or_default(),
                    block_time: block_time__,
                })
            }
        }
        deserializer.deserialize_struct("babylon.epoching.v1.ValStateUpdate", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Validator {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.addr.is_empty() {
            len += 1;
        }
        if self.power != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.epoching.v1.Validator", len)?;
        if !self.addr.is_empty() {
            struct_ser.serialize_field("addr", pbjson::private::base64::encode(&self.addr).as_str())?;
        }
        if self.power != 0 {
            struct_ser.serialize_field("power", ToString::to_string(&self.power).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Validator {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "addr",
            "power",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Addr,
            Power,
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
                            "addr" => Ok(GeneratedField::Addr),
                            "power" => Ok(GeneratedField::Power),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Validator;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.epoching.v1.Validator")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<Validator, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut addr__ = None;
                let mut power__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::Addr => {
                            if addr__.is_some() {
                                return Err(serde::de::Error::duplicate_field("addr"));
                            }
                            addr__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Power => {
                            if power__.is_some() {
                                return Err(serde::de::Error::duplicate_field("power"));
                            }
                            power__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(Validator {
                    addr: addr__.unwrap_or_default(),
                    power: power__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.epoching.v1.Validator", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ValidatorLifecycle {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.val_addr.is_empty() {
            len += 1;
        }
        if !self.val_life.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.epoching.v1.ValidatorLifecycle", len)?;
        if !self.val_addr.is_empty() {
            struct_ser.serialize_field("valAddr", &self.val_addr)?;
        }
        if !self.val_life.is_empty() {
            struct_ser.serialize_field("valLife", &self.val_life)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ValidatorLifecycle {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "val_addr",
            "valAddr",
            "val_life",
            "valLife",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValAddr,
            ValLife,
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
                            "valAddr" | "val_addr" => Ok(GeneratedField::ValAddr),
                            "valLife" | "val_life" => Ok(GeneratedField::ValLife),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ValidatorLifecycle;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.epoching.v1.ValidatorLifecycle")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<ValidatorLifecycle, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut val_addr__ = None;
                let mut val_life__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::ValAddr => {
                            if val_addr__.is_some() {
                                return Err(serde::de::Error::duplicate_field("valAddr"));
                            }
                            val_addr__ = Some(map.next_value()?);
                        }
                        GeneratedField::ValLife => {
                            if val_life__.is_some() {
                                return Err(serde::de::Error::duplicate_field("valLife"));
                            }
                            val_life__ = Some(map.next_value()?);
                        }
                    }
                }
                Ok(ValidatorLifecycle {
                    val_addr: val_addr__.unwrap_or_default(),
                    val_life: val_life__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.epoching.v1.ValidatorLifecycle", FIELDS, GeneratedVisitor)
    }
}
