package middleware

import (
	"encoding/base64"
	"log/slog"

	"github.com/dnonakolesax/noted-notes/internal/consts"
	"github.com/dnonakolesax/noted-notes/internal/rnd"
	"github.com/valyala/fasthttp"
)

const requestIDSize = 16

func CommonMW(requestHandler fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Access-Control-Allow-Origin", consts.URL + "/files/*")
		ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET,PUT, PATCH, DELETE, OPTIONS")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
		ctx.Request.SetUserValue(consts.CtxUserIDKey, "00000000-0000-0000-0000-000000000000")
		requestID := ctx.Request.Header.Peek("X-Request-Id")
		var reqID string
		if requestID == nil {
			requestID = rnd.NotSafeGenRandomString(requestIDSize)
			reqID = base64.RawURLEncoding.EncodeToString(requestID)
			ctx.Request.Header.Set("X-Request-Id", reqID)
		} else {
			reqID = string(requestID)
		}
		slog.Info("Received Request",
			slog.String("method", string(ctx.Method())),
			slog.String("path", string(ctx.Path())),
			slog.String("ip", ctx.RemoteIP().String()),
			slog.String("requestId", reqID),
			slog.String("userAgent", string(ctx.UserAgent())),
		)
		now := ctx.Time().UnixMilli()
		requestHandler(ctx)
		end := ctx.Time().UnixMilli()
		slog.Info("Completed request",
			slog.String("method", string(ctx.Method())),
			slog.String("path", string(ctx.Path())),
			slog.String("requestId", reqID),
			slog.Int("status", ctx.Response.StatusCode()),
			slog.Int("duration", int(end-now)),
		)
		ctx.Response.Header.Add(fasthttp.HeaderContentType, "application/json")
	}
}
