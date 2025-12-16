package model

type CodeBlock struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

type BlockVO struct {
	FileID   string `json:"kernel_id,omitempty"`
	ID       string `json:"id,omitempty"`
	PrevID   string `json:"prev_id,omitempty"`
	Language string `json:"language,omitempty"`
}

type NewBlockRDTO struct {
	ID string `json:"id"`
}