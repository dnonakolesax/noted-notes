package model

type Access struct {
	UserID string `json:"user_id"`
	Level string `json:"level"`
}

type RevokeDTO struct {
	UserID string `json:"user_id"`
}
