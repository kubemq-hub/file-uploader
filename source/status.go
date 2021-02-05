package source

type Status struct {
	Waiting    int `json:"waiting"`
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
}
