package entity

// SignRequest is an internal container for the request to sign the text with key (SHA512)
type SignRequest struct {
	Text string `json:"text,omitempty"`
	Key  string `json:"key,omitempty"`
}

// SignResponse contains result signature in Hex format
type SignResponse struct {
	Hex string `json:"hex"`
}
