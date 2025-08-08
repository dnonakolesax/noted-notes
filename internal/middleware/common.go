package middleware

import (
	"github.com/valyala/fasthttp"
)

func CommonMW(requestHandler fasthttp.RequestHandler) fasthttp.RequestHandler {
    return func(ctx *fasthttp.RequestCtx) {
        requestHandler(ctx)
		ctx.Response.Header.Add(fasthttp.HeaderContentType, "application/json")
    }
}
