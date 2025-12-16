package services

import (
	"context"
	"github.com/dnonakolesax/noted-notes/internal/model"
)

type BlockRepository interface {
	Get(ctx context.Context, id string) ([]byte, error)
	DeleteS3(id string) error
	Move(id string, newParent string) error
	Add(block model.BlockVO) error
	Delete(id string) error
	Upload(ctx context.Context, id string, text []byte) error
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

func (bs *BlockService) Delete(id string) error {
	err := bs.repo.Delete(id)

	if err != nil {
		return err
	}

	err = bs.repo.DeleteS3(id)

	if err != nil {
		return err
	}

	return nil
}

func (bs *BlockService) Move(id string, parentID string) error {
	err := bs.repo.Move(id, parentID)

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
