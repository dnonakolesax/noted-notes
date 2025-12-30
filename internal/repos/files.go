package repos

import (
	"context"
	"fmt"

	"github.com/dnonakolesax/noted-notes/internal/db/sql"
	"github.com/dnonakolesax/noted-notes/internal/model"
	"github.com/google/uuid"
)

const GET_FILE_NAME string = "get_file"
const DELETE_FILE_NAME string = "delete_file"
const RENAME_FILE_NAME string = "rename_file"

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
		return model.FileVO{}, fmt.Errorf("request returned no rows")
	}
	err = resp.Scan(&file.Owner, &file.Public, &file.Author, &file.BlocksIds, &file.BlocksLanguages, &file.BlocksPrevs)

	if err != nil {
		return model.FileVO{}, err
	}

	if resp.Next() {
		return model.FileVO{}, fmt.Errorf("request returned more than one row")
	}
	err = resp.Close()
	if err != nil {
		return model.FileVO{}, err
	}
	return file, nil
}

func (fr *filesRepo) DeleteFile(fileId uuid.UUID) (error) {
	err := fr.dbWorker.Exec(context.TODO(), fr.dbWorker.Requests[DELETE_FILE_NAME], fileId)
	if err != nil {
		return err
	}
	return nil
}

func (fr *filesRepo) RenameFile(fileId string) (error) {
	err := fr.dbWorker.Exec(context.TODO(), fr.dbWorker.Requests[RENAME_FILE_NAME], fileId)
	if err != nil {
		return err
	}
	return nil
}

func NewFilesRepo(dbWorker sql.PGXWorker) *filesRepo {
	return &filesRepo{
		dbWorker: dbWorker,
	}
}
