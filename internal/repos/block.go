package repos

import (
	"context"

	"github.com/dnonakolesax/noted-notes/internal/s3"
)

type BlockRepo struct {
	worker s3.S3Worker
}

func NewBlockRepo(worker s3.S3Worker) *BlockRepo {
	return &BlockRepo{
		worker: worker,
	}
}

func (br *BlockRepo) GetBlock(id string) ([]byte, error) {
	fileName := "block_" + id

	bytes, err := br.worker.DownloadFile(context.TODO(), "noted", fileName)

	if err != nil {
		return nil, err
	}
	return bytes, nil
}
