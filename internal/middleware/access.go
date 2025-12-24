package middleware

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/dnonakolesax/noted-notes/internal/consts"
	"github.com/valyala/fasthttp"
)

type AccessService interface {
	Get(fileID string, userID string, byBlock bool) (string, error)
}

type AccessMW struct {
	service AccessService
}

func NewAccessMW(service AccessService) *AccessMW {
	return &AccessMW{service: service}
}

func (am *AccessMW) GetRights(ctx *fasthttp.RequestCtx) (string, error) {
	byBlock := false
	fileID := ctx.Request.UserValue("fileID")
	if fileID == nil {
		dirID := ctx.Request.UserValue("dirID")
		if dirID == nil {
			blockID := ctx.Request.UserValue("blockID")
			if blockID == nil {
				slog.Warn("fileid and dirid is nil")
				ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
				return "", fmt.Errorf("fileid and dirid is nil")
			}
			fileID = blockID
			byBlock = true
		} else {
			fileID = dirID
		}
	}
	userID := ctx.Request.UserValue(consts.CtxUserIDKey)
	if userID == nil {
		slog.Warn("userid is nil")
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
		return "", fmt.Errorf("userid is nil")
	}
	access, err := am.service.Get(fileID.(string), userID.(string), byBlock)

	if err != nil {
		slog.Warn("err getting access", slog.String("error", err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
		return "", fmt.Errorf("err getting access: %s", err.Error())
	}
	ctx.Request.SetUserValue("access", access)

	return access, nil
}

func (am *AccessMW) Write(requestHandler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		rights, err := am.GetRights(ctx)

		if err != nil {
			return
		}

		if !(strings.Contains(rights, "w")) {
			slog.Warn("bad rights: not writer", slog.String("rights", rights), slog.String("userid", ctx.Request.UserValue(consts.CtxUserIDKey).(string)))
			ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
			return
		}

		requestHandler(ctx)
	}
}

func (am *AccessMW) Read(requestHandler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		rights, err := am.GetRights(ctx)

		if err != nil {
			return
		}

		if !(strings.Contains(rights, "r")) {
			slog.Warn("bad rights: not reader", slog.String("rights", rights), slog.String("userid", ctx.Request.UserValue(consts.CtxUserIDKey).(string)))
			ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
			return
		}

		requestHandler(ctx)
	}
}

func (am *AccessMW) Own(requestHandler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		rights, err := am.GetRights(ctx)

		if err != nil {
			return
		}

		if !(strings.Contains(rights, "o")) {
			slog.Warn("bad rights: not owner", slog.String("rights", rights), slog.String("userid", ctx.Request.UserValue(consts.CtxUserIDKey).(string)))
			ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
			return
		}

		requestHandler(ctx)
	}
}
