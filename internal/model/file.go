package model

type FileDTO struct {
	Blocks     []CodeBlock `json:"blocks"`
	Owner      string      `json:"owner"`
	LastUpdate string      `json:"lastUpdate"`
	Rights     string      `json:"rights"`
}

type FileVO struct {
	BlocksIds       []string `json:"blocksIds"`
	BlocksLanguages []string `json:"blocksLanguages"`
	BlocksPrevs     []string `json:"blocksprevs"`
	Owner           string   `json:"owner"`
	Public          bool     `json:"public"`
}
