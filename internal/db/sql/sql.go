package sql

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RDBConfig struct {
	Address  string
	Port     int
	DBName   string
	Login    string
	Password string
}

type PGXConn struct {
	pool   *pgxpool.Pool
}

type RDBErr struct {
	Type  string
	Field string
}

func (err RDBErr) Error() string {
	return err.Type + " " + err.Field
}

type PGXResponse struct {
	rows pgx.Rows
}

func NewPGXConn(config RDBConfig) (*PGXConn, error) {
	var err error
	pool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		config.Login,
		config.Password,
		config.Address,
		config.Port,
		config.DBName))
	if err != nil {
		return nil, err
	}
	return &PGXConn{pool: pool}, nil
}

func (pc *PGXConn) Disconnect() {
	pc.pool.Close()
}

type PGXWorker struct {
	Conn *PGXConn
	Requests map[string]string
}

func LoadSQLRequests(dirPath string) (map[string]string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %v", err)
	}

	sqlRequests := make(map[string]string)

	for _, file := range files {
		if file.IsDir() {
			continue // Пропускаем директории
		}

		if filepath.Ext(file.Name()) != ".sql" {
			continue // Пропускаем файлы без .sql расширения
		}

		filePath := filepath.Join(dirPath, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file.Name(), err)
		}

		// Получаем имя файла без расширения
		fileName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		sqlRequests[fileName] = string(content)
	}

	return sqlRequests, nil
}

func NewPGXWorker(conn *PGXConn) (*PGXWorker, error) {
	requests, err := LoadSQLRequests("./internal/db/sql/requests")

	if err != nil {
		return nil, err
	}

	return &PGXWorker{
		Conn: conn,
		Requests: requests,
	}, nil
}

func (pw *PGXWorker) Exec(ctx context.Context, sql string, args ...interface{}) error {
	_, err := pw.Conn.pool.Exec(ctx, sql, args...)

	// var pgErr *pgconn.PgError

	// if errors.As(err, &pgErr) {
	// 	rdbErr := new(RDBErr)
	// 	rdbErr.Type = pgErr.Code
	// 	rdbErr.Field = pgErr.ColumnName
	// 	return rdbErr
	// }

	if err != nil {
		fmt.Printf("db error exec: %v\n", err)
		return err
	}

	return nil
}

func (pw *PGXWorker) Query(ctx context.Context, sql string, args ...interface{}) (*PGXResponse, error) {
	result, err := pw.Conn.pool.Query(ctx, sql, args...)

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return &PGXResponse{}, RDBErr{Type: pgErr.Code, Field: pgErr.ColumnName}
	}

	return &PGXResponse{result}, nil
}

type PgTXR struct {
	Request string
	Data []any
}

func (pw *PGXWorker) Transaction(ctx context.Context, request []PgTXR) error {
	 tx, err := pw.Conn.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    
    defer tx.Rollback(ctx)

	for _, r := range request {
		_, err = tx.Exec(ctx, r.Request, r.Data...)
		if err != nil {
			return fmt.Errorf("tx error: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
    	return fmt.Errorf("commit transaction: %w", err)
    }
	return nil
}

func (pr *PGXResponse) Next() bool {
	return pr.rows.Next()
}

// func (pr *PGXResponse) Size() int {
// 	return len(pr.rows.)
// }

func (pr *PGXResponse) Scan(dest ...any) error {

	// for pr.rows.Next() {
		err := pr.rows.Scan(dest...)
		if err != nil {
			return fmt.Errorf("scan error: %v", err)
		}
	//}

	return nil
}

func (pr *PGXResponse) Close() error {
	pr.rows.Close()
	//Err() on the returned Rows must be checked after the Rows is closed to determine if the query executed successfully as some errors can only be detected by reading the entire response.
	//e.g. A divide by zero error on the last row.
	err := pr.rows.Err()
	if err != nil {
		return err
	}
	return nil
}
