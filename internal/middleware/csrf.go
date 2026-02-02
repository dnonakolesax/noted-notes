package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/dnonakolesax/noted-notes/internal/consts"
	"github.com/dnonakolesax/noted-notes/internal/cookies"
	"github.com/dnonakolesax/noted-notes/internal/rnd"
	"github.com/dnonakolesax/noted-notes/internal/xerrors"
	"github.com/valyala/fasthttp"
)

type csrfRepo interface {
	Get(sessID string) (string, error)
	Continue(sessID string) error
	Set(sessID string, token string)
}

type CSRFMW struct {
	repo   csrfRepo
	logger *slog.Logger
}

func NewCSRFMW(repo csrfRepo) *CSRFMW {
	return &CSRFMW{
		repo:   repo,
		logger: slog.Default(),
	}
}

func isSafe(method string) bool {
	return (method == "GET") || (method == "HEAD") || (method == "OPTIONS")
}

func (cm *CSRFMW) safeMW(ctx *fasthttp.RequestCtx, sid string, contex context.Context) error {
	token, err := rnd.GenRandomString(64)

	if err != nil {
		cm.logger.ErrorContext(contex, "rnd gen error", slog.String(consts.ErrorLoggerKey, err.Error()))
		return err
	}

	cm.repo.Set(sid, token)
	cookies.SetupCSRF(ctx, token)
	return nil
}

func (cm *CSRFMW) unsafeMW(ctx *fasthttp.RequestCtx, token string, contex context.Context) error {
	csrft := ctx.Request.Header.Cookie(consts.CSRFCookieKey)

	if csrft == nil {
		cm.logger.WarnContext(contex, "csrf token not passed")
		return fmt.Errorf("cstf token not passed")
	}
	if token != string(csrft) {
		cm.logger.WarnContext(contex, "csrf: tokens not equal")
		return fmt.Errorf("csrf: tokens not equal")
	}
	return nil
}

func (cm *CSRFMW) MW(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		at := ctx.Request.Header.Cookie(consts.ATCookieKey)
		trace := string(ctx.Request.Header.Peek(consts.HTTPHeaderXRequestID))
		contex := context.WithValue(context.Background(), consts.TraceContextKey, trace)

		if at != nil {
			sid := ctx.UserValue(consts.CtxSessIDKey)

			if sid == nil {
				cm.logger.WarnContext(contex, "csrf: sid not found in user value")
				ctx.SetStatusCode(fasthttp.StatusUnauthorized)
				return
			}

			token, err := cm.repo.Get(sid.(string))

			if isSafe(string(ctx.Method())) {
				if err != nil {
					if errors.Is(err, xerrors.ErrCSRFTokenNotFound) {
						err = cm.safeMW(ctx, sid.(string), contex)
						if err != nil {
							ctx.SetStatusCode(fasthttp.StatusInternalServerError)
							return
						}
					} else {
						cm.logger.ErrorContext(contex, "error agetting csrf")
						ctx.SetStatusCode(fasthttp.StatusUnauthorized)
						return
					}
				} else {
					cookies.SetupCSRF(ctx, token)
				}
			} else {
				if err != nil {
					cm.logger.ErrorContext(contex, "error getting csrf")
					ctx.SetStatusCode(fasthttp.StatusUnauthorized)
					return
				}
				err = cm.unsafeMW(ctx, token, contex)

				if err != nil {
					ctx.SetStatusCode(fasthttp.StatusUnauthorized)
					return
				}
			}
		} else if !(isSafe(string(ctx.Method()))) {
			cm.logger.WarnContext(contex, "no at passed")
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			return
		}

		h(ctx)
	})
}
