package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/dnonakolesax/noted-notes/internal/consts"
	"github.com/dnonakolesax/noted-notes/internal/middleware"
	"github.com/dnonakolesax/noted-notes/internal/model"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type AccService interface {
	GetAll(fileID string) ([]model.Access, error)
	Grant(fileID string, userID string, level string) error
	Update(fileID string, userID string, level string) error
	Revoke(fileID string, userID string) error
}

type AccessHandler struct {
	service  AccService
	accessMW *middleware.AccessMW
}

func NewAccessHandler(service AccService, accessMW *middleware.AccessMW) *AccessHandler {
	return &AccessHandler{
		service:  service,
		accessMW: accessMW,
	}
}

func (ah *AccessHandler) Grant(ctx *fasthttp.RequestCtx) {
	fileID := ctx.UserValue("fileID").(string)

	var payload model.Access
	err := json.Unmarshal(ctx.Request.Body(), &payload)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	err = ah.service.Grant(fileID, payload.UserID, payload.Level)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (ah *AccessHandler) Update(ctx *fasthttp.RequestCtx) {
	fileID := ctx.UserValue("fileID").(string)
	
	var payload model.Access
	err := json.Unmarshal(ctx.Request.Body(), &payload)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	err = ah.service.Update(fileID, payload.UserID, payload.Level)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (ah *AccessHandler) Revoke(ctx *fasthttp.RequestCtx) {
	fileID := ctx.UserValue("fileID").(string)	
	var payload model.Access
	err := json.Unmarshal(ctx.Request.Body(), &payload)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	err = ah.service.Revoke(fileID, payload.UserID)
	
	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (ah *AccessHandler) Get(ctx *fasthttp.RequestCtx) {
	fileID := ctx.UserValue("fileID").(string)
	
	accessList, err := ah.service.GetAll(fileID)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	listJson, err := json.Marshal(accessList)

	if err != nil {
		fmt.Printf("Error marshaling: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetBody(listJson)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (ah *AccessHandler) RegisterRoutes(r *router.Group) {
	group := r.Group("/access")
	group.OPTIONS("/{fileID}", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Access-Control-Allow-Origin", consts.URL + "/access/*")
		ctx.Response.Header.Add("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.GET("/{fileID}", ah.accessMW.Own(ah.Get))
	group.POST("/{fileID}", ah.accessMW.Own(ah.Grant))
	group.PUT("/{fileID}", ah.accessMW.Own(ah.Update))
	group.DELETE("/{fileID}", ah.accessMW.Own(ah.Revoke))
}
