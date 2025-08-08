package services

import (
	"github.com/dnonakolesax/noted-notes/internal/model"
	"github.com/google/uuid"
)

type filesRepo interface {
	GetFile(fileId string, userId string) (model.FileVO, error)
}

type blocksRepo interface {
	GetBlock(blockId string) ([]byte, error)
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
		block, err := fs.brepo.GetBlock(trueBlockId)
		if err != nil {
			return model.FileDTO{}, err
		}
		blocks[idx] = model.CodeBlock{
			Code: string(block),
			Language: fileVO.BlocksLanguages[idx],
		}
	}
	
	return model.FileDTO{
		Blocks: blocks,
		Owner: fileVO.Owner,
	}, nil
}

func (fs *FilesService) Remove() error {
	return nil
}

func NewFilesService(frepo filesRepo, brepo blocksRepo) *FilesService {
	return &FilesService{
		frepo: frepo,
		brepo: brepo,
	}
}
