package common

import (
	"encoding/json"
	"fmt"
)

// BytesToType converts array of bytes to the variable of type T and returns pointer to it
func BytesToType[T any](b []byte) (*T, error) {
	var t T
	err := json.Unmarshal(b, &t)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %s", err)
	}
	return &t, nil
}
