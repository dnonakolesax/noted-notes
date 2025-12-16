package model

type FileDTO struct {
	Blocks     []CodeBlock `json:"blocks"`
	Owner      string      `json:"owner"`
	LastUpdate string      `json:"lastUpdate"`
}

type FileVO struct {
	BlocksIds       []string `json:"blocksIds"`
	BlocksLanguages []string `json:"blocksLanguages"`
	Owner           string   `json:"owner"`
	Public          bool     `json:"public"`
}
