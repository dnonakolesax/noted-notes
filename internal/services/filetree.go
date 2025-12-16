package services

import (
	"log/slog"

	"github.com/google/uuid"
)

type TreeRepo interface {
	Add(fileName string, uuid string, isDir bool, parentDir string) error
	Rename(uuid string, newName string) error
	Move(uuid string, newParent string) error
}

type TreeService struct {
	repo TreeRepo
}

func NewTreeService(repo TreeRepo) *TreeService {
	return &TreeService{repo: repo}
}

func (ts *TreeService) Add(filename string, uuid uuid.UUID, isDir bool, parentDir uuid.UUID) error {
	fileID := uuid.String()
	parentID := parentDir.String()

	err := ts.repo.Add(filename, fileID, isDir, parentID)

	if err != nil {
		slog.Error("error repo add file", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (ts *TreeService) Rename(fileUUID uuid.UUID, newName string) error {
	fileID := fileUUID.String()

	err := ts.repo.Rename(fileID, newName)

	if err != nil {
		slog.Error("error repo rename file", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (ts *TreeService) Move(fileUUID uuid.UUID, newParent uuid.UUID) error {
	fileID := fileUUID.String()
	newParentID := newParent.String() 

	err := ts.repo.Move(fileID, newParentID)

	if err != nil {
		slog.Error("error repo rename file", slog.String("error", err.Error()))
		return err
	}
	return nil
}

func (ts *TreeService) ChangePrivacy(userId uuid.UUID,id uuid.UUID, isPublic bool) error {
	return nil
}
func (ts *TreeService) GrantAccess(userId uuid.UUID, id uuid.UUID, targetUserId uuid.UUID, accessType string) error {
	return nil
}
