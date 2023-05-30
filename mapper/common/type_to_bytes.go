package common

import (
	"encoding/json"
	"fmt"
)

func TypeToBytes[T any](t *T) ([]byte, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal: %s", err)
	}
	return b, nil
}
