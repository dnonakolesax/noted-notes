package repos

import (
	"context"
	"log/slog"

	"github.com/dnonakolesax/noted-notes/internal/db/sql"
	"github.com/dnonakolesax/noted-notes/internal/model"
)

type AccessRepo struct {
	dbWorker *sql.PGXWorker
}

const SELECT_ACCESS_NAME string = "select_access"
const LIST_ACCESS_NAME string = "list_access"
const SELECT_ACCESS_BLOCK_NAME string = "select_access_block"
const GRANT_ACCESS_NAME string = "grant_access"
const UPDATE_ACCESS_NAME string = "update_access"
const REVOKE_ACCESS_NAME string = "revoke_access"

func NewAccessRepo(worker *sql.PGXWorker) *AccessRepo {
	return &AccessRepo{dbWorker: worker}
}

func (ar *AccessRepo) Get(fileID string, userID string, byBlock bool) (string, error) {
	requestName := SELECT_ACCESS_NAME

	if byBlock {
		requestName = SELECT_ACCESS_BLOCK_NAME
	}

	res, err := ar.dbWorker.Query(context.TODO(), ar.dbWorker.Requests[requestName], fileID, userID)
	slog.Info("ac req done")
	if err != nil {
		slog.Info(err.Error())
		return "", err
	}
	defer res.Close()

	if !res.Next() {
		slog.Info("no next")
		return "", nil
	}

	var owner string
	var acStr string
	var isPublic bool

	err = res.Scan(&owner, &acStr, &isPublic)

	if err != nil {
		return "", err
	}

	if owner == userID {
		return "orwx", nil
	}

	if acStr == "" && isPublic {
		slog.Info("access is now r")
		acStr = "r"
	}
	slog.Info(acStr)

	return acStr, nil
}

func (ar *AccessRepo) GetAll(fileID string) ([]model.Access, error) {
	resp, err := ar.dbWorker.Query(context.TODO(), ar.dbWorker.Requests[LIST_ACCESS_NAME], fileID)

	if err != nil {
		return []model.Access{}, err
	}

	accessList := make([]model.Access, 0)
	for resp.Next() {
		var currAcc model.Access
		err = resp.Scan(&currAcc.UserID, &currAcc.Level)
		if err != nil {
			return []model.Access{}, err
		}	
		accessList = append(accessList, currAcc)
	}

	err = resp.Close()
	if err != nil {
		return []model.Access{}, err
	}
	return accessList, nil
}


func (ar *AccessRepo) Grant(fileID string, userID string, level string) error {
	err := ar.dbWorker.Exec(context.Background(), ar.dbWorker.Requests[GRANT_ACCESS_NAME], fileID, userID, level)

	if err != nil {
		return err
	}
	
	return nil
}

func (ar *AccessRepo) Update(fileID string, userID string, level string) error {
	err := ar.dbWorker.Exec(context.Background(), ar.dbWorker.Requests[UPDATE_ACCESS_NAME], fileID, userID, level)

	if err != nil {
		return err
	}

	return nil
}

func (ar *AccessRepo) Revoke(fileID string, userID string) error {
	err := ar.dbWorker.Exec(context.Background(), ar.dbWorker.Requests[REVOKE_ACCESS_NAME], fileID, userID)

	if err != nil {
		return err
	}

	return nil
}