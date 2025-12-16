package model

type FileInTreeDTO struct {
	Name      string `json:"name,omitempty"`
	IsDir     bool   `json:"is_dir,omitempty"`
	ParentDir string `json:"parent_dir,omitempty"`
}

type FileInTreeRDTO struct {
	ID string `json:"id,omitempty"`
}
