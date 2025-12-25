package services

import (
	"context"
	"fmt"

	"github.com/dnonakolesax/noted-notes/internal/model"
)

type BlockRepository interface {
	Get(ctx context.Context, id string) ([]byte, error)
	DeleteS3(id string) error
	Move(id1 string, id2 string) error
	Add(block model.BlockVO) error
	Delete(id string, fileID string) error
	Upload(ctx context.Context, id string, text []byte) error
	VerifyMove(id1 string, id2 string, fileID string) (bool, error)
}

type BlockService struct {
	repo BlockRepository
}

func (bs *BlockService) Save(block model.BlockVO) error {
	err := bs.repo.Upload(context.Background(), block.ID, []byte{})

	if err != nil {
		return err
	}

	err = bs.repo.Add(block)

	if err != nil {
		return err
	}

	return nil
}

func (bs *BlockService) Delete(id string, fileID string) error {
	err := bs.repo.Delete(id, fileID)

	if err != nil {
		return err
	}

	err = bs.repo.DeleteS3(id)

	if err != nil {
		return err
	}

	return nil
}

func (bs *BlockService) Move(id1 string, id2 string, fileID string, direction string) error {
	switch direction {
	case "up":
		tmp := id1
		id1 = id2
		id2 = tmp
	case "down":
	default:
		return fmt.Errorf("invalid direction: %s", direction)
	}
	ok, err := bs.repo.VerifyMove(id2, id1, fileID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("invalid move operation from %s to %s", id1, id2)
	}
	err = bs.repo.Move(id1, id2)
	if err != nil {
		return err
	}
	return nil
}

func NewBlockService(repo BlockRepository) *BlockService {
	return &BlockService{
		repo: repo,
	}
}
