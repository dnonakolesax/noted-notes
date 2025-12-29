package repos

import (
	"context"
	"log/slog"

	"github.com/dnonakolesax/noted-notes/internal/db/sql"
)

type AccessRepo struct {
	dbWorker *sql.PGXWorker
}

const SELECT_ACCESS_NAME string = "select_access"
const SELECT_ACCESS_BLOCK_NAME string = "select_access_block"

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
