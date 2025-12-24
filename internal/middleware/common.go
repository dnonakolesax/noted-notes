package middleware

import (
	"github.com/dnonakolesax/noted-notes/internal/consts"
	"github.com/valyala/fasthttp"
)

func CommonMW(requestHandler fasthttp.RequestHandler) fasthttp.RequestHandler {
    return func(ctx *fasthttp.RequestCtx) {
        ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
        ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
        ctx.Request.SetUserValue(consts.CtxUserIDKey, "8ef6dafb-1143-4da7-9654-144a696a140e")
        requestHandler(ctx)
		ctx.Response.Header.Add(fasthttp.HeaderContentType, "application/json")
    }
}
