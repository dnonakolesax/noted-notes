package cookies

import (
	"time"

	"github.com/dnonakolesax/noted-notes/internal/consts"
	"github.com/valyala/fasthttp"
)

func SetupAccessCookies(ctx *fasthttp.RequestCtx, at string, rt string, it string) {
	atCookie := fasthttp.Cookie{}
	atCookie.SetKey(consts.ATCookieKey)
	atCookie.SetValue(at)
	atCookie.SetMaxAge(int((time.Minute*20).Seconds()))
	atCookie.SetHTTPOnly(true)
	atCookie.SetSecure(true)
	atCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	atCookie.SetPath("/")

	rtCookie := fasthttp.Cookie{}
	rtCookie.SetKey(consts.RTCookieKey)
	rtCookie.SetValue(rt)
	rtCookie.SetMaxAge(int((time.Hour*24*50).Seconds()))
	rtCookie.SetHTTPOnly(true)
	rtCookie.SetSecure(true)
	rtCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	rtCookie.SetPath("/")

	idtCookie := fasthttp.Cookie{}
	idtCookie.SetKey(consts.IDTCookieKey)
	idtCookie.SetValue(it)
	idtCookie.SetHTTPOnly(true)
	idtCookie.SetSecure(true)
	idtCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	idtCookie.SetPath("/")

	ctx.Response.Header.SetCookie(&atCookie)
	ctx.Response.Header.SetCookie(&rtCookie)
	ctx.Response.Header.SetCookie(&idtCookie)
	ctx.Request.Header.SetCookie(consts.ATCookieKey, at)
	ctx.Request.Header.SetCookie(consts.RTCookieKey, rt)
	ctx.Request.Header.SetCookie(consts.IDTCookieKey, it)
}

func SetupCSRF (ctx *fasthttp.RequestCtx, csrf string) {
	csrfCookie := fasthttp.Cookie{}
	csrfCookie.SetKey(consts.CSRFCookieKey)
	csrfCookie.SetValue(csrf)
	csrfCookie.SetMaxAge(int((time.Minute*20).Seconds()))
	csrfCookie.SetHTTPOnly(true)
	csrfCookie.SetSecure(true)
	csrfCookie.SetSameSite(fasthttp.CookieSameSiteLaxMode)
	csrfCookie.SetPath("/")

	ctx.Response.Header.SetCookie(&csrfCookie)
	ctx.Request.Header.SetCookie(consts.CSRFCookieKey, csrf)
}
