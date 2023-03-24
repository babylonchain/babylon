package grandpa_types

import (
	"encoding/json"
)

type U32 uint32

// NewU32 creates a new U32 type
func NewU32(u uint32) U32 {
	return U32(u)
}

// UnmarshalJSON fills u with the JSON encoded byte array given by b
func (u *U32) UnmarshalJSON(b []byte) error {
	var tmp uint32
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	*u = U32(tmp)
	return nil
}

// MarshalJSON returns a JSON encoded byte array of u
func (u U32) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint32(u))
}
