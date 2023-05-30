package entity

// AddUserRequest is an internal container for the request to add a new row with the user data to the postgres repo
type AddUserRequest struct {
	Name string `json:"name,omitempty"`
	Age  int    `json:"age,omitempty"`
}

// AddUserResponse is an internal container for the response that contains the id of the row where data was inserted
type AddUserResponse struct {
	Id int64 `json:"id"`
}
