// @generated
impl serde::Serialize for ChainInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.chain_id.is_empty() {
            len += 1;
        }
        if self.latest_header.is_some() {
            len += 1;
        }
        if self.latest_forks.is_some() {
            len += 1;
        }
        if self.timestamped_headers_count != 0 {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.zoneconcierge.v1.ChainInfo", len)?;
        if !self.chain_id.is_empty() {
            struct_ser.serialize_field("chainId", &self.chain_id)?;
        }
        if let Some(v) = self.latest_header.as_ref() {
            struct_ser.serialize_field("latestHeader", v)?;
        }
        if let Some(v) = self.latest_forks.as_ref() {
            struct_ser.serialize_field("latestForks", v)?;
        }
        if self.timestamped_headers_count != 0 {
            struct_ser.serialize_field("timestampedHeadersCount", ToString::to_string(&self.timestamped_headers_count).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ChainInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "chain_id",
            "chainId",
            "latest_header",
            "latestHeader",
            "latest_forks",
            "latestForks",
            "timestamped_headers_count",
            "timestampedHeadersCount",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChainId,
            LatestHeader,
            LatestForks,
            TimestampedHeadersCount,
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
                            "chainId" | "chain_id" => Ok(GeneratedField::ChainId),
                            "latestHeader" | "latest_header" => Ok(GeneratedField::LatestHeader),
                            "latestForks" | "latest_forks" => Ok(GeneratedField::LatestForks),
                            "timestampedHeadersCount" | "timestamped_headers_count" => Ok(GeneratedField::TimestampedHeadersCount),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ChainInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.zoneconcierge.v1.ChainInfo")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<ChainInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut chain_id__ = None;
                let mut latest_header__ = None;
                let mut latest_forks__ = None;
                let mut timestamped_headers_count__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::ChainId => {
                            if chain_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("chainId"));
                            }
                            chain_id__ = Some(map.next_value()?);
                        }
                        GeneratedField::LatestHeader => {
                            if latest_header__.is_some() {
                                return Err(serde::de::Error::duplicate_field("latestHeader"));
                            }
                            latest_header__ = map.next_value()?;
                        }
                        GeneratedField::LatestForks => {
                            if latest_forks__.is_some() {
                                return Err(serde::de::Error::duplicate_field("latestForks"));
                            }
                            latest_forks__ = map.next_value()?;
                        }
                        GeneratedField::TimestampedHeadersCount => {
                            if timestamped_headers_count__.is_some() {
                                return Err(serde::de::Error::duplicate_field("timestampedHeadersCount"));
                            }
                            timestamped_headers_count__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(ChainInfo {
                    chain_id: chain_id__.unwrap_or_default(),
                    latest_header: latest_header__,
                    latest_forks: latest_forks__,
                    timestamped_headers_count: timestamped_headers_count__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.zoneconcierge.v1.ChainInfo", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for Forks {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.headers.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.zoneconcierge.v1.Forks", len)?;
        if !self.headers.is_empty() {
            struct_ser.serialize_field("headers", &self.headers)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for Forks {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "headers",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            Headers,
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
                            "headers" => Ok(GeneratedField::Headers),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = Forks;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.zoneconcierge.v1.Forks")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<Forks, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut headers__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::Headers => {
                            if headers__.is_some() {
                                return Err(serde::de::Error::duplicate_field("headers"));
                            }
                            headers__ = Some(map.next_value()?);
                        }
                    }
                }
                Ok(Forks {
                    headers: headers__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.zoneconcierge.v1.Forks", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for IndexedHeader {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.chain_id.is_empty() {
            len += 1;
        }
        if !self.hash.is_empty() {
            len += 1;
        }
        if self.height != 0 {
            len += 1;
        }
        if self.babylon_header.is_some() {
            len += 1;
        }
        if self.babylon_epoch != 0 {
            len += 1;
        }
        if !self.babylon_tx_hash.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.zoneconcierge.v1.IndexedHeader", len)?;
        if !self.chain_id.is_empty() {
            struct_ser.serialize_field("chainId", &self.chain_id)?;
        }
        if !self.hash.is_empty() {
            struct_ser.serialize_field("hash", pbjson::private::base64::encode(&self.hash).as_str())?;
        }
        if self.height != 0 {
            struct_ser.serialize_field("height", ToString::to_string(&self.height).as_str())?;
        }
        if let Some(v) = self.babylon_header.as_ref() {
            struct_ser.serialize_field("babylonHeader", v)?;
        }
        if self.babylon_epoch != 0 {
            struct_ser.serialize_field("babylonEpoch", ToString::to_string(&self.babylon_epoch).as_str())?;
        }
        if !self.babylon_tx_hash.is_empty() {
            struct_ser.serialize_field("babylonTxHash", pbjson::private::base64::encode(&self.babylon_tx_hash).as_str())?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for IndexedHeader {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "chain_id",
            "chainId",
            "hash",
            "height",
            "babylon_header",
            "babylonHeader",
            "babylon_epoch",
            "babylonEpoch",
            "babylon_tx_hash",
            "babylonTxHash",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ChainId,
            Hash,
            Height,
            BabylonHeader,
            BabylonEpoch,
            BabylonTxHash,
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
                            "chainId" | "chain_id" => Ok(GeneratedField::ChainId),
                            "hash" => Ok(GeneratedField::Hash),
                            "height" => Ok(GeneratedField::Height),
                            "babylonHeader" | "babylon_header" => Ok(GeneratedField::BabylonHeader),
                            "babylonEpoch" | "babylon_epoch" => Ok(GeneratedField::BabylonEpoch),
                            "babylonTxHash" | "babylon_tx_hash" => Ok(GeneratedField::BabylonTxHash),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = IndexedHeader;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.zoneconcierge.v1.IndexedHeader")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<IndexedHeader, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut chain_id__ = None;
                let mut hash__ = None;
                let mut height__ = None;
                let mut babylon_header__ = None;
                let mut babylon_epoch__ = None;
                let mut babylon_tx_hash__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::ChainId => {
                            if chain_id__.is_some() {
                                return Err(serde::de::Error::duplicate_field("chainId"));
                            }
                            chain_id__ = Some(map.next_value()?);
                        }
                        GeneratedField::Hash => {
                            if hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("hash"));
                            }
                            hash__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::Height => {
                            if height__.is_some() {
                                return Err(serde::de::Error::duplicate_field("height"));
                            }
                            height__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BabylonHeader => {
                            if babylon_header__.is_some() {
                                return Err(serde::de::Error::duplicate_field("babylonHeader"));
                            }
                            babylon_header__ = map.next_value()?;
                        }
                        GeneratedField::BabylonEpoch => {
                            if babylon_epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("babylonEpoch"));
                            }
                            babylon_epoch__ = 
                                Some(map.next_value::<::pbjson::private::NumberDeserialize<_>>()?.0)
                            ;
                        }
                        GeneratedField::BabylonTxHash => {
                            if babylon_tx_hash__.is_some() {
                                return Err(serde::de::Error::duplicate_field("babylonTxHash"));
                            }
                            babylon_tx_hash__ = 
                                Some(map.next_value::<::pbjson::private::BytesDeserialize<_>>()?.0)
                            ;
                        }
                    }
                }
                Ok(IndexedHeader {
                    chain_id: chain_id__.unwrap_or_default(),
                    hash: hash__.unwrap_or_default(),
                    height: height__.unwrap_or_default(),
                    babylon_header: babylon_header__,
                    babylon_epoch: babylon_epoch__.unwrap_or_default(),
                    babylon_tx_hash: babylon_tx_hash__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.zoneconcierge.v1.IndexedHeader", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ProofEpochSealed {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if !self.validator_set.is_empty() {
            len += 1;
        }
        if self.proof_epoch_info.is_some() {
            len += 1;
        }
        if self.proof_epoch_val_set.is_some() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.zoneconcierge.v1.ProofEpochSealed", len)?;
        if !self.validator_set.is_empty() {
            struct_ser.serialize_field("validatorSet", &self.validator_set)?;
        }
        if let Some(v) = self.proof_epoch_info.as_ref() {
            struct_ser.serialize_field("proofEpochInfo", v)?;
        }
        if let Some(v) = self.proof_epoch_val_set.as_ref() {
            struct_ser.serialize_field("proofEpochValSet", v)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ProofEpochSealed {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "validator_set",
            "validatorSet",
            "proof_epoch_info",
            "proofEpochInfo",
            "proof_epoch_val_set",
            "proofEpochValSet",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ValidatorSet,
            ProofEpochInfo,
            ProofEpochValSet,
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
                            "validatorSet" | "validator_set" => Ok(GeneratedField::ValidatorSet),
                            "proofEpochInfo" | "proof_epoch_info" => Ok(GeneratedField::ProofEpochInfo),
                            "proofEpochValSet" | "proof_epoch_val_set" => Ok(GeneratedField::ProofEpochValSet),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ProofEpochSealed;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.zoneconcierge.v1.ProofEpochSealed")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<ProofEpochSealed, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut validator_set__ = None;
                let mut proof_epoch_info__ = None;
                let mut proof_epoch_val_set__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::ValidatorSet => {
                            if validator_set__.is_some() {
                                return Err(serde::de::Error::duplicate_field("validatorSet"));
                            }
                            validator_set__ = Some(map.next_value()?);
                        }
                        GeneratedField::ProofEpochInfo => {
                            if proof_epoch_info__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proofEpochInfo"));
                            }
                            proof_epoch_info__ = map.next_value()?;
                        }
                        GeneratedField::ProofEpochValSet => {
                            if proof_epoch_val_set__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proofEpochValSet"));
                            }
                            proof_epoch_val_set__ = map.next_value()?;
                        }
                    }
                }
                Ok(ProofEpochSealed {
                    validator_set: validator_set__.unwrap_or_default(),
                    proof_epoch_info: proof_epoch_info__,
                    proof_epoch_val_set: proof_epoch_val_set__,
                })
            }
        }
        deserializer.deserialize_struct("babylon.zoneconcierge.v1.ProofEpochSealed", FIELDS, GeneratedVisitor)
    }
}
impl serde::Serialize for ProofFinalizedChainInfo {
    #[allow(deprecated)]
    fn serialize<S>(&self, serializer: S) -> std::result::Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        use serde::ser::SerializeStruct;
        let mut len = 0;
        if self.proof_tx_in_block.is_some() {
            len += 1;
        }
        if self.proof_header_in_epoch.is_some() {
            len += 1;
        }
        if self.proof_epoch_sealed.is_some() {
            len += 1;
        }
        if !self.proof_epoch_submitted.is_empty() {
            len += 1;
        }
        let mut struct_ser = serializer.serialize_struct("babylon.zoneconcierge.v1.ProofFinalizedChainInfo", len)?;
        if let Some(v) = self.proof_tx_in_block.as_ref() {
            struct_ser.serialize_field("proofTxInBlock", v)?;
        }
        if let Some(v) = self.proof_header_in_epoch.as_ref() {
            struct_ser.serialize_field("proofHeaderInEpoch", v)?;
        }
        if let Some(v) = self.proof_epoch_sealed.as_ref() {
            struct_ser.serialize_field("proofEpochSealed", v)?;
        }
        if !self.proof_epoch_submitted.is_empty() {
            struct_ser.serialize_field("proofEpochSubmitted", &self.proof_epoch_submitted)?;
        }
        struct_ser.end()
    }
}
impl<'de> serde::Deserialize<'de> for ProofFinalizedChainInfo {
    #[allow(deprecated)]
    fn deserialize<D>(deserializer: D) -> std::result::Result<Self, D::Error>
    where
        D: serde::Deserializer<'de>,
    {
        const FIELDS: &[&str] = &[
            "proof_tx_in_block",
            "proofTxInBlock",
            "proof_header_in_epoch",
            "proofHeaderInEpoch",
            "proof_epoch_sealed",
            "proofEpochSealed",
            "proof_epoch_submitted",
            "proofEpochSubmitted",
        ];

        #[allow(clippy::enum_variant_names)]
        enum GeneratedField {
            ProofTxInBlock,
            ProofHeaderInEpoch,
            ProofEpochSealed,
            ProofEpochSubmitted,
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
                            "proofTxInBlock" | "proof_tx_in_block" => Ok(GeneratedField::ProofTxInBlock),
                            "proofHeaderInEpoch" | "proof_header_in_epoch" => Ok(GeneratedField::ProofHeaderInEpoch),
                            "proofEpochSealed" | "proof_epoch_sealed" => Ok(GeneratedField::ProofEpochSealed),
                            "proofEpochSubmitted" | "proof_epoch_submitted" => Ok(GeneratedField::ProofEpochSubmitted),
                            _ => Err(serde::de::Error::unknown_field(value, FIELDS)),
                        }
                    }
                }
                deserializer.deserialize_identifier(GeneratedVisitor)
            }
        }
        struct GeneratedVisitor;
        impl<'de> serde::de::Visitor<'de> for GeneratedVisitor {
            type Value = ProofFinalizedChainInfo;

            fn expecting(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
                formatter.write_str("struct babylon.zoneconcierge.v1.ProofFinalizedChainInfo")
            }

            fn visit_map<V>(self, mut map: V) -> std::result::Result<ProofFinalizedChainInfo, V::Error>
                where
                    V: serde::de::MapAccess<'de>,
            {
                let mut proof_tx_in_block__ = None;
                let mut proof_header_in_epoch__ = None;
                let mut proof_epoch_sealed__ = None;
                let mut proof_epoch_submitted__ = None;
                while let Some(k) = map.next_key()? {
                    match k {
                        GeneratedField::ProofTxInBlock => {
                            if proof_tx_in_block__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proofTxInBlock"));
                            }
                            proof_tx_in_block__ = map.next_value()?;
                        }
                        GeneratedField::ProofHeaderInEpoch => {
                            if proof_header_in_epoch__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proofHeaderInEpoch"));
                            }
                            proof_header_in_epoch__ = map.next_value()?;
                        }
                        GeneratedField::ProofEpochSealed => {
                            if proof_epoch_sealed__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proofEpochSealed"));
                            }
                            proof_epoch_sealed__ = map.next_value()?;
                        }
                        GeneratedField::ProofEpochSubmitted => {
                            if proof_epoch_submitted__.is_some() {
                                return Err(serde::de::Error::duplicate_field("proofEpochSubmitted"));
                            }
                            proof_epoch_submitted__ = Some(map.next_value()?);
                        }
                    }
                }
                Ok(ProofFinalizedChainInfo {
                    proof_tx_in_block: proof_tx_in_block__,
                    proof_header_in_epoch: proof_header_in_epoch__,
                    proof_epoch_sealed: proof_epoch_sealed__,
                    proof_epoch_submitted: proof_epoch_submitted__.unwrap_or_default(),
                })
            }
        }
        deserializer.deserialize_struct("babylon.zoneconcierge.v1.ProofFinalizedChainInfo", FIELDS, GeneratedVisitor)
    }
}
