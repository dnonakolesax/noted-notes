package repos

import (
	"context"

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

	if (byBlock) {
		requestName = SELECT_ACCESS_BLOCK_NAME
	}

	res, err := ar.dbWorker.Query(context.TODO(), ar.dbWorker.Requests[requestName], fileID, userID)
	if err != nil {
		return "", err
	}
	defer res.Close()

	if !res.Next() {
		return "", nil
	}

	var owner string
	var acStr string

	err = res.Scan(&owner, &acStr)

	if err != nil {
		return "", err
	}

	if owner == userID {
		return "orwx", nil
	}

	return acStr, nil
}
