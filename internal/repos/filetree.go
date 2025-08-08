package repos

import (
	"context"
	"fmt"

	dbsql "github.com/dnonakolesax/noted-notes/internal/db/sql"
)

const INSERT_FILE_NAME string = "insert_file"
const RENAME_FILE_NAME string = "rename_file"
const MOVE_FILE_NAME string = "move_file"
const CHANGE_PRIVACY_NAME string = "change_privacy"
const GRANT_ACCESS_NAME string = "grant_access"
const GET_ACCESS_NAME string = "get_access"

type FileTreeRepo struct {
	worker *dbsql.PGXWorker
}

func NewFileTreeRepo(worker *dbsql.PGXWorker) *FileTreeRepo {
	return &FileTreeRepo{
		worker: worker,
	}
}

func (fr *FileTreeRepo) AddFile(fileName string, uuid string, isDir bool, parentDir string) (error) {
	err := fr.worker.Exec(context.TODO(), fr.worker.Requests[INSERT_FILE_NAME], fileName, uuid, isDir, parentDir)

	if err != nil {
		return err
	}
	return nil
}

func (fr *FileTreeRepo) RenameFile(uuid string, newName string) (error) {
	err := fr.worker.Exec(context.TODO(), fr.worker.Requests[RENAME_FILE_NAME], uuid, newName)

	if err != nil {
		return err
	}
	return nil
}

func (fr *FileTreeRepo) MoveFile(uuid string, newParent string) (error) {
	err := fr.worker.Exec(context.TODO(), fr.worker.Requests[MOVE_FILE_NAME], uuid, newParent)

	if err != nil {
		return err
	}
	return nil
}

func (fr *FileTreeRepo) ChangePrivacy(uuid string, isPrivate bool) (error) {
	err := fr.worker.Exec(context.TODO(), fr.worker.Requests[CHANGE_PRIVACY_NAME], uuid, isPrivate)

	if err != nil {
		return err
	}
	return nil
}

func (fr *FileTreeRepo) GrantAccess(uuid string, userUuid string, access string) (error) {
	err := fr.worker.Exec(context.TODO(), fr.worker.Requests[GRANT_ACCESS_NAME], uuid, userUuid, access)

	if err != nil {
		return err
	}
	return nil
}

func (fr *FileTreeRepo) GetAccess(uuid string) (string, error) {
	res, err := fr.worker.Query(context.TODO(), fr.worker.Requests[GET_ACCESS_NAME], uuid)

	if err != nil {
		return "", err
	}

	var access string
	err = res.Scan(&access)
	if err != nil {
		return "", err
	}

	if res.Next() {
		return "", fmt.Errorf("Multiple access for file")
	}


	err = res.Close()
	if err != nil {
		return "", err
	}

	return access, nil
}