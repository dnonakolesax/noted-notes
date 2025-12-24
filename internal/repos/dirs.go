package repos

import (
	"context"

	"github.com/dnonakolesax/noted-notes/internal/db/sql"
	"github.com/dnonakolesax/noted-notes/internal/model"
)

const GET_DIR_NAME string = "get_dir"

type dirsRepo struct {
	dbWorker sql.PGXWorker
}

func (dr *dirsRepo) Get(fileId string, userId string) ([]model.Directory, error) {
	resp, err := dr.dbWorker.Query(context.TODO(), dr.dbWorker.Requests[GET_DIR_NAME], fileId, userId)

	if err != nil {
		return []model.Directory{}, err
	}

	dirs := make([]model.Directory, 0)
	for resp.Next() {
		var currDir model.Directory
		var currOwner string
		err = resp.Scan(&currDir.FileId, &currDir.Name, &currDir.Dir, &currDir.Rights, &currOwner)
		if currOwner == userId {
			currDir.Rights = "orwx"
		}
		if err != nil {
			return []model.Directory{}, err
		}	
		dirs = append(dirs, currDir)
	}

	err = resp.Close()
	if err != nil {
		return []model.Directory{}, err
	}
	return dirs, nil
}

func (dr *dirsRepo) Remove() error {
	return nil
}

func NewDirsRepo(dbWorker sql.PGXWorker) *dirsRepo {
	return &dirsRepo{
		dbWorker: dbWorker,
	}
}
