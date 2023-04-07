// Go Substrate RPC Client (GSRPC) provides APIs and types around Polkadot and any Substrate-based chain RPC calls
//
// Copyright 2019 Centrifuge GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grandpa_types

import (
	"fmt"
)

// H256 is a hash containing 256 bits (32 bytes), typically used in blocks, extrinsics and as a sane default
type (
	H256 [32]byte
	Hash H256
)

// NewHash creates a new Hash type
func NewHash(b []byte) Hash {
	h := Hash{}
	copy(h[:], b)
	return h
}

// NewH256 creates a new H256 type
func NewH256(b []byte) H256 {
	h := H256{}
	copy(h[:], b)
	return h
}

// Hex returns a hex string representation of the value (not of the encoded value)
func (h H256) Hex() string {
	return fmt.Sprintf("%#x", h[:])
}
