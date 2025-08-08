package repos

import (
	"context"
	"fmt"

	"github.com/dnonakolesax/noted-notes/internal/db/sql"
	"github.com/dnonakolesax/noted-notes/internal/model"
)

const GET_FILE_NAME string = "get_file"

type filesRepo struct {
	dbWorker sql.PGXWorker
}

func (fr *filesRepo) GetFile(fileId string, userId string) (model.FileVO, error) {
	resp, err := fr.dbWorker.Query(context.TODO(), fr.dbWorker.Requests[GET_FILE_NAME], fileId)

	if err != nil {
		return model.FileVO{}, err
	}

	file := model.FileVO{}
	if !resp.Next() {
		return model.FileVO{}, fmt.Errorf("Request returned no rows")
	}
	err = resp.Scan(&file.Owner, &file.BlocksIds, &file.BlocksLanguages)

	if err != nil {
		return model.FileVO{}, err
	}

	if resp.Next() {
		return model.FileVO{}, fmt.Errorf("Request returned more than one row")
	}
	err = resp.Close()
	if err != nil {
		return model.FileVO{}, err
	}
	return file, nil
}

func NewFilesRepo(dbWorker sql.PGXWorker) *filesRepo {
	return &filesRepo{
		dbWorker: dbWorker,
	}
}
