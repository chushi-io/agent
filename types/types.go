package types

type RunStatusUpdate struct {
	Add     int    `json:"add,omitempty"`
	Change  int    `json:"change,omitempty"`
	Destroy int    `json:"destroy,omitempty"`
	Status  string `json:"status,omitempty"`
}

type PlanStatusUpdate struct {
	Status string `json:"status,omitempty"`
}

type ApplyStatusUpdate struct {
	Status string `json:"status,omitempty"`
}
