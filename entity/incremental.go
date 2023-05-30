package entity

// IncrementRequest is an internal container for the request to increment Key for Value in redis
type IncrementRequest struct {
	Key   string `json:"key,omitempty"`
	Value int64  `json:"value,omitempty"`
}

// IncrementResponse is and internal container for the incrementing results
type IncrementResponse struct {
	Value int64 `json:"value,omitempty"`
}
