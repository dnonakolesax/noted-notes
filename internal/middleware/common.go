package middleware

import (
	"github.com/valyala/fasthttp"
)

func CommonMW(requestHandler fasthttp.RequestHandler) fasthttp.RequestHandler {
    return func(ctx *fasthttp.RequestCtx) {
        ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
        ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
        requestHandler(ctx)
		ctx.Response.Header.Add(fasthttp.HeaderContentType, "application/json")
    }
}
