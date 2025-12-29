package repos

import (
	"errors"

	"github.com/dnonakolesax/noted-notes/internal/consts"
	"github.com/dnonakolesax/noted-notes/internal/xerrors"
	"github.com/muesli/cache2go"
)

type CSRF struct {
	cache *cache2go.CacheTable
}

func NewCSRF() *CSRF {
	return &CSRF{
		cache: cache2go.Cache("csrf"),
	}
}

func (c *CSRF) Set(sessID string, token string) {
	c.cache.Add(sessID, consts.ATLifetime, token)
}

func (c *CSRF) Get(sessID string) (string, error) {
	item, err := c.cache.Value(sessID)

	if err != nil {
		if errors.Is(err, cache2go.ErrKeyNotFoundOrLoadable) || errors.Is(err, cache2go.ErrKeyNotFound) {
			return "", xerrors.ErrCSRFTokenNotFound
		}
		return "", err
	}

	return item.Data().(string), nil
}

func (c *CSRF) Continue(sessID string) error {
	item, err := c.cache.Value(sessID)

	if err != nil {
		if errors.Is(err, cache2go.ErrKeyNotFoundOrLoadable) {
			return xerrors.ErrCSRFTokenNotFound
		}
		return err
	}
	item.KeepAlive()
	return nil
}
