package services

import (
	"github.com/dnonakolesax/noted-notes/internal/model"
	"github.com/google/uuid"
)

type dirsRepo interface {
	Get(fileId string, userId string) ([]model.Directory, error)
}

type DirectoryService struct {
	repo  dirsRepo
}

func (ds *DirectoryService) Get(fileId uuid.UUID, userId uuid.UUID) ([]model.Directory, error) {
	dirs, err := ds.repo.Get(fileId.String(), userId.String())

	if err != nil {
		return []model.Directory{}, err
	}

	return dirs, nil
}

func (ds *DirectoryService) Remove() error {
	return nil
}

func NewDirsService(drepo dirsRepo) *DirectoryService {
	return &DirectoryService{
		repo:  drepo,
	}
}
