package consts

import "time"

const (
	CtxUserIDKey = "user_id"
	CtxSessIDKey = "session_id"
)

const (
	ATCookieKey  = "NTD-DNAnAT"
	RTCookieKey  = "NTD-DNART"
	IDTCookieKey = "NTD-DNALT"
)

const (
	ATLifetime = time.Minute*25
)

const (
	CSRFCookieKey = "NTD-DNaCT"
)

type ContextKey string

const (
	TraceContextKey ContextKey = "trace"
	TraceLoggerKey  string     = "trace-id"
)

const (
	ErrorLoggerKey = "error"
)

const (
	HTTPHeaderXRequestID = "X-Request-Id"
)

const (
	URL = "https://dnk33.com"
)
