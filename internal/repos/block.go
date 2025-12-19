package repos

import (
	"context"
	"fmt"
	"strings"

	"github.com/dnonakolesax/noted-notes/internal/db/sql"
	"github.com/dnonakolesax/noted-notes/internal/model"
	"github.com/dnonakolesax/noted-notes/internal/s3"
)

type BlockRepo struct {
	worker   s3.S3Worker
	dbWorker sql.PGXWorker
}

const MOVE_BLOCK_NAME1 string = "move_block_1"
const MOVE_BLOCK_NAME11 string = "move_block_11"
const MOVE_BLOCK_NAME2 string = "move_block_2"
const ADD_BLOCK_NAME string = "add_block"
const DELETE_BLOCK_NAME1 string = "delete_block_1"
const DELETE_BLOCK_NAME2 string = "delete_block_2"
const DELETE_ALL_BLOCKS_NAME string = "delete_all_blocks"

func NewBlockRepo(worker s3.S3Worker, dbWorker sql.PGXWorker) *BlockRepo {
	return &BlockRepo{
		worker:   worker,
		dbWorker: dbWorker,
	}
}

func (br *BlockRepo) Get(ctx context.Context, id string) ([]byte, error) {
	fileName := "block_" + id
	fmt.Println(fileName)
	bytes, err := br.worker.DownloadFile(ctx, "noted", fileName)

	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (br *BlockRepo) DeleteS3(id string) error {
	fileName := "block_" + strings.ReplaceAll(id, "-", "")
	
	err := br.worker.MoveS3Object(context.TODO(), "noted", fileName, "noted-icecold", fileName)

	if err != nil {
		return err
	}
	return nil
}

func (br *BlockRepo) DeleteAll(fileID string) error {
	res, err := br.dbWorker.Query(context.TODO(), DELETE_ALL_BLOCKS_NAME, fileID)

	if err != nil {
		return err
	}
	defer res.Close()
	for res.Next() {
		var id string
		err = res.Scan(&id)

		if err != nil {
			return err
		}

		fname := "block_" + id
		err = br.worker.MoveS3Object(context.Background(), "noted", fname, "noted-icecold", fname)
		if err != nil {
			return err
		}
	}
	return nil
}

func (br *BlockRepo) Upload(ctx context.Context, id string, text []byte) error {
	fileName := "block_" + id
	err := br.worker.UploadFile(ctx, "noted", fileName, text)
	
	if err != nil {
		return err
	}
	return nil
}

func (br *BlockRepo) Move(id string, newParent string, direction string) error {
	var r1 sql.PgTXR
	if direction == "down" {
		r1 = sql.PgTXR{
			Request: br.dbWorker.Requests[MOVE_BLOCK_NAME1],
			Data: []any{id},
		}
	} else {
		r1 = sql.PgTXR{
			Request: br.dbWorker.Requests[MOVE_BLOCK_NAME11],
			Data: []any{id, newParent},
		}
	}
	r2 := sql.PgTXR{
		Request: br.dbWorker.Requests[MOVE_BLOCK_NAME2],
		Data: []any{id, newParent},
	}
	err := br.dbWorker.Transaction(context.TODO(), []sql.PgTXR{r1, r2})

	if err != nil {
		return err
	}
	return nil
}

func (br *BlockRepo) Add(block model.BlockVO) error {
	fmt.Println(block.ID)
	var err error
	if block.PrevID == "" {
		err = br.dbWorker.Exec(context.TODO(), br.dbWorker.Requests[ADD_BLOCK_NAME],
			block.ID, block.FileID, nil, block.Language)
	} else {
		err = br.dbWorker.Exec(context.TODO(), br.dbWorker.Requests[ADD_BLOCK_NAME],
			block.ID, block.FileID, block.PrevID, block.Language)
	}

	if err != nil {
		return err
	}
	return nil
}

func (br *BlockRepo) Delete(id string) error {
	r1 := sql.PgTXR{
		Request: br.dbWorker.Requests[DELETE_BLOCK_NAME1],
		Data: []any{id},
	}
	r2 := sql.PgTXR{
		Request: br.dbWorker.Requests[DELETE_BLOCK_NAME2],
		Data: []any{id},
	}
	err := br.dbWorker.Transaction(context.TODO(), []sql.PgTXR{r1, r2})

	if err != nil {
		return err
	}
	return nil
}
