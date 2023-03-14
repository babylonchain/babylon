// @generated
impl serde::Serialize for BlsSig {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.epoch_num != 0 {
            len += 1;
        }
        if !self.last_commit_hash.is_empty() {
            len += 1;
        }
        if !self.bls_sig.is_empty() {
            len += 1;
        }
        if !self.signer_address.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.checkpointing.v1.BlsSig", len)?;
        if self.epoch_num != 0 {
            struct_ser.serialize_field("epochNum", ToString::to_string(&self.epoch_num).as_str())?;
        }
        if !self.last_commit_hash.is_empty() {
            struct_ser.serialize_field("lastCommitHash", pbjson::private::base64::encode(&self.last_commit_hash).as_str())?;
        }
        if !self.bls_sig.is_empty() {
            struct_ser.serialize_field("blsSig", pbjson::private::base64::encode(&self.bls_sig).as_str())?;
        }
        if !self.signer_address.is_empty() {
            struct_ser.serialize_field("signerAddress", &self.signer_address)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for BlsSig {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "epoch_num",
            "epochNum",
            "last_commit_hash",
            "lastCommitHash",
            "bls_sig",
            "blsSig",
            "signer_address",
            "signerAddress",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EpochNum,
            LastCommitHash,
            BlsSig,
            SignerAddress,
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
                            "epochNum" | "epoch_num" => Ok(GeneratedField::EpochNum),
                            "lastCommitHash" | "last_commit_hash" => Ok(GeneratedField::LastCommitHash),
                            "blsSig" | "bls_sig" => Ok(GeneratedField::BlsSig),
                            "signerAddress" | "signer_address" => Ok(GeneratedField::SignerAddress),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = BlsSig;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.checkpointing.v1.BlsSig")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<BlsSig, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut epoch_num__ = None;
                let mut last_commit_hash__ = None;
                let mut bls_sig__ = None;
                let mut signer_address__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::EpochNum => {
                            if epoch_num__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochNum"));
                            }
                            epoch_num__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastCommitHash => {
                            if last_commit_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastCommitHash"));
                            }
                            last_commit_hash__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlsSig => {
                            if bls_sig__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blsSig"));
                            }
                            bls_sig__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::SignerAddress => {
                            if signer_address__.is_some() {
                                return Err(serde::de::Error::duplicate_field("signerAddress"));
                            }
                            signer_address__ = Some(map.next_value()?);
                        }
                    }
                }
                Ok(BlsSig {
                    epoch_num: epoch_num__.unwrap_or_default(),
                    last_commit_hash: last_commit_hash__.unwrap_or_default(),
                    bls_sig: bls_sig__.unwrap_or_default(),
                    signer_address: signer_address__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.checkpointing.v1.BlsSig", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for CheckpointStateUpdate {
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
        let mut struct_ser = serializer.serialize_struct("babylon.checkpointing.v1.CheckpointStateUpdate", len)?;
        if self.state != 0 {
            let v = CheckpointStatus::from_i32(self.state)
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
impl<'de> serde::Deserialize<'de> for CheckpointStateUpdate {
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
            type Value = CheckpointStateUpdate;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.checkpointing.v1.CheckpointStateUpdate")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<CheckpointStateUpdate, V::Error>
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
                            state__ = Some(map.next_value::<CheckpointStatus>()? as i32);
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
                Ok(CheckpointStateUpdate {
                    state: state__.unwrap_or_default(),
                    block_height: block_height__.unwrap_or_default(),
                    block_time: block_time__,
                })
            }
        }
        deserializer.deserialize_struct("babylon.checkpointing.v1.CheckpointStateUpdate", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for CheckpointStatus {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        let variant = match self {
            Self::CkptStatusAccumulating => "CKPT_STATUS_ACCUMULATING",
            Self::CkptStatusSealed => "CKPT_STATUS_SEALED",
            Self::CkptStatusSubmitted => "CKPT_STATUS_SUBMITTED",
            Self::CkptStatusConfirmed => "CKPT_STATUS_CONFIRMED",
            Self::CkptStatusFinalized => "CKPT_STATUS_FINALIZED",
        };
        serializer.serialize_str(variant)
    }
}
impl<'de> serde::Deserialize<'de> for CheckpointStatus {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "CKPT_STATUS_ACCUMULATING",
            "CKPT_STATUS_SEALED",
            "CKPT_STATUS_SUBMITTED",
            "CKPT_STATUS_CONFIRMED",
            "CKPT_STATUS_FINALIZED",
        ];

        struct GeneratedVisitor;

        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = CheckpointStatus;

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
                    .and_then(CheckpointStatus::from_i32)
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
                    .and_then(CheckpointStatus::from_i32)
                    .ok_or_else(|| {
                        serde::de::Error::invalid_value(serde::de::Unexpected::Unsigned(v), &self)
                    })
            }

            fn visit_str<E>(self, value: &str) -> std::result::Result<Self::Value, E>
            where
                E: serde::de::Error,
            {
                match value {
                    "CKPT_STATUS_ACCUMULATING" => Ok(CheckpointStatus::CkptStatusAccumulating),
                    "CKPT_STATUS_SEALED" => Ok(CheckpointStatus::CkptStatusSealed),
                    "CKPT_STATUS_SUBMITTED" => Ok(CheckpointStatus::CkptStatusSubmitted),
                    "CKPT_STATUS_CONFIRMED" => Ok(CheckpointStatus::CkptStatusConfirmed),
                    "CKPT_STATUS_FINALIZED" => Ok(CheckpointStatus::CkptStatusFinalized),
                    _ => Err(serde::de::Error::unknown_variant(value, FIELDS)),
                }
            }
        }
        deserializer.deserialize_any(GeneratedVisitor)
    }
}
impl serde::Serialize for RawCheckpoint {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.epoch_num != 0 {
            len += 1;
        }
        if !self.last_commit_hash.is_empty() {
            len += 1;
        }
        if !self.bitmap.is_empty() {
            len += 1;
        }
        if !self.bls_multi_sig.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.checkpointing.v1.RawCheckpoint", len)?;
        if self.epoch_num != 0 {
            struct_ser.serialize_field("epochNum", ToString::to_string(&self.epoch_num).as_str())?;
        }
        if !self.last_commit_hash.is_empty() {
            struct_ser.serialize_field("lastCommitHash", pbjson::private::base64::encode(&self.last_commit_hash).as_str())?;
        }
        if !self.bitmap.is_empty() {
            struct_ser.serialize_field("bitmap", pbjson::private::base64::encode(&self.bitmap).as_str())?;
        }
        if !self.bls_multi_sig.is_empty() {
            struct_ser.serialize_field("blsMultiSig", pbjson::private::base64::encode(&self.bls_multi_sig).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for RawCheckpoint {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "epoch_num",
            "epochNum",
            "last_commit_hash",
            "lastCommitHash",
            "bitmap",
            "bls_multi_sig",
            "blsMultiSig",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            EpochNum,
            LastCommitHash,
            Bitmap,
            BlsMultiSig,
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
                            "epochNum" | "epoch_num" => Ok(GeneratedField::EpochNum),
                            "lastCommitHash" | "last_commit_hash" => Ok(GeneratedField::LastCommitHash),
                            "bitmap" => Ok(GeneratedField::Bitmap),
                            "blsMultiSig" | "bls_multi_sig" => Ok(GeneratedField::BlsMultiSig),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RawCheckpoint;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.checkpointing.v1.RawCheckpoint")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<RawCheckpoint, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut epoch_num__ = None;
                let mut last_commit_hash__ = None;
                let mut bitmap__ = None;
                let mut bls_multi_sig__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::EpochNum => {
                            if epoch_num__.is_some() {
                                return Err(serde::de::Error::duplicate_field("epochNum"));
                            }
                            epoch_num__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::LastCommitHash => {
                            if last_commit_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lastCommitHash"));
                            }
                            last_commit_hash__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Bitmap => {
                            if bitmap__.is_some() {
                                return Err(serde::de::Error::duplicate_field("bitmap"));
                            }
                            bitmap__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BlsMultiSig => {
                            if bls_multi_sig__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blsMultiSig"));
                            }
                            bls_multi_sig__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(RawCheckpoint {
                    epoch_num: epoch_num__.unwrap_or_default(),
                    last_commit_hash: last_commit_hash__.unwrap_or_default(),
                    bitmap: bitmap__.unwrap_or_default(),
                    bls_multi_sig: bls_multi_sig__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.checkpointing.v1.RawCheckpoint", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for RawCheckpointWithMeta {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.ckpt.is_some() {
            len += 1;
        }
        if self.status != 0 {
            len += 1;
        }
        if !self.bls_aggr_pk.is_empty() {
            len += 1;
        }
        if self.power_sum != 0 {
            len += 1;
        }
        if !self.lifecycle.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.checkpointing.v1.RawCheckpointWithMeta", len)?;
        if let Some(v) = self.ckpt.as_ref() {
            struct_ser.serialize_field("ckpt", v)?;
        }
        if self.status != 0 {
            let v = CheckpointStatus::from_i32(self.status)
                .ok_or_else(|| serde::ser::Error::custom(format!("Invalid variant {}", self.status)))?;
            struct_ser.serialize_field("status", &v)?;
        }
        if !self.bls_aggr_pk.is_empty() {
            struct_ser.serialize_field("blsAggrPk", pbjson::private::base64::encode(&self.bls_aggr_pk).as_str())?;
        }
        if self.power_sum != 0 {
            struct_ser.serialize_field("powerSum", ToString::to_string(&self.power_sum).as_str())?;
        }
        if !self.lifecycle.is_empty() {
            struct_ser.serialize_field("lifecycle", &self.lifecycle)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for RawCheckpointWithMeta {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "ckpt",
            "status",
            "bls_aggr_pk",
            "blsAggrPk",
            "power_sum",
            "powerSum",
            "lifecycle",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Ckpt,
            Status,
            BlsAggrPk,
            PowerSum,
            Lifecycle,
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
                            "ckpt" => Ok(GeneratedField::Ckpt),
                            "status" => Ok(GeneratedField::Status),
                            "blsAggrPk" | "bls_aggr_pk" => Ok(GeneratedField::BlsAggrPk),
                            "powerSum" | "power_sum" => Ok(GeneratedField::PowerSum),
                            "lifecycle" => Ok(GeneratedField::Lifecycle),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = RawCheckpointWithMeta;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.checkpointing.v1.RawCheckpointWithMeta")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<RawCheckpointWithMeta, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut ckpt__ = None;
                let mut status__ = None;
                let mut bls_aggr_pk__ = None;
                let mut power_sum__ = None;
                let mut lifecycle__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::Ckpt => {
                            if ckpt__.is_some() {
                                return Err(serde::de::Error::duplicate_field("ckpt"));
                            }
                            ckpt__ = map.next_value()?;
                        }
                        GeneratedField::Status => {
                            if status__.is_some() {
                                return Err(serde::de::Error::duplicate_field("status"));
                            }
                            status__ = Some(map.next_value::<CheckpointStatus>()? as i32);
                        }
                        GeneratedField::BlsAggrPk => {
                            if bls_aggr_pk__.is_some() {
                                return Err(serde::de::Error::duplicate_field("blsAggrPk"));
                            }
                            bls_aggr_pk__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::PowerSum => {
                            if power_sum__.is_some() {
                                return Err(serde::de::Error::duplicate_field("powerSum"));
                            }
                            power_sum__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Lifecycle => {
                            if lifecycle__.is_some() {
                                return Err(serde::de::Error::duplicate_field("lifecycle"));
                            }
                            lifecycle__ = Some(map.next_value()?);
                        }
                    }
                }
                Ok(RawCheckpointWithMeta {
                    ckpt: ckpt__,
                    status: status__.unwrap_or_default(),
                    bls_aggr_pk: bls_aggr_pk__.unwrap_or_default(),
                    power_sum: power_sum__.unwrap_or_default(),
                    lifecycle: lifecycle__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.checkpointing.v1.RawCheckpointWithMeta", FIELDS, GeneratedVisitor)
    }
}
