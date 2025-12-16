package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/dnonakolesax/noted-notes/internal/model"
	"github.com/google/uuid"
)

type filesRepo interface {
	GetFile(fileId string, userId string) (model.FileVO, error)
	DeleteFile(fileId uuid.UUID) (error)
}

type blocksRepo interface {
	Get(ctx context.Context, blockId string) ([]byte, error)
	DeleteAll(fileID string) error 
}

type FilesService struct {
	frepo filesRepo
	brepo blocksRepo
}

func (fs *FilesService) Get(fileId uuid.UUID, userId uuid.UUID) (model.FileDTO, error) {
	fileVO, err := fs.frepo.GetFile(fileId.String(), userId.String())

	if err != nil {
		return model.FileDTO{}, err
	} 

	blocks := make([]model.CodeBlock, len(fileVO.BlocksIds))
	
	for idx, blockId := range fileVO.BlocksIds {
		var trueBlockId string
		for _, b := range blockId {
			if b != '-' {
				trueBlockId += string(b)
			}
		}
		block, err := fs.brepo.Get(context.Background(), trueBlockId)
		if err != nil {
			return model.FileDTO{}, err
		}
		blocks[idx] = model.CodeBlock{
			Code: string(block),
			Language: fileVO.BlocksLanguages[idx],
		}

		if fileVO.Public {
			path := fmt.Sprintf("%s/%s/%s", "/noted/codes/kernels", 
									fileId.String(), userId.String())	

			err := os.MkdirAll(path, os.ModeDir)

			if err != nil {
				slog.Error("error create dir", "error", err.Error())
				return model.FileDTO{}, err
			}

			blockFile, err := os.Create(path + "/block_" + blockId)
			
			if err != nil {
				slog.Error("error create block", "error", err.Error())
				return model.FileDTO{}, err
			}

			_, err = blockFile.Write(block)

			if err != nil {
				slog.Error("error write block", "error", err.Error())
				return model.FileDTO{}, err
			}
		}
	}
	
	return model.FileDTO{
		Blocks: blocks,
		Owner: fileVO.Owner,
	}, nil
}

func (fs *FilesService) Delete(fileID uuid.UUID) error {
	err := fs.frepo.DeleteFile(fileID)

	if err != nil {
		slog.Error("error remove file", "error", err.Error())
		return err
	}

	err = fs.brepo.DeleteAll(fileID.String())

	if err != nil {
		slog.Error("error remove blocks", "error", err.Error())
		return err
	}

	return nil
}

func NewFilesService(frepo filesRepo, brepo blocksRepo) *FilesService {
	return &FilesService{
		frepo: frepo,
		brepo: brepo,
	}
}
