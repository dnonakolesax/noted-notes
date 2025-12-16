package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/dnonakolesax/noted-notes/internal/middleware"
	"github.com/dnonakolesax/noted-notes/internal/model"
	"github.com/fasthttp/router"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

type DirectoriesService interface {
	Get(fileId uuid.UUID, userId uuid.UUID) ([]model.Directory, error)
	Remove() error
}

type DirectoriesHandler struct {
	service DirectoriesService
}

func (dh *DirectoriesHandler) Get(ctx *fasthttp.RequestCtx) {
	fileId := ctx.UserValue("dirId").(string)

	userId := "0416603d-9a5c-4290-a1dd-62babfea991e"
	fileUUID, err := uuid.Parse(fileId)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	userUUID, err := uuid.Parse(userId)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	dirs, err := dh.service.Get(fileUUID, userUUID)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	dirJson, err := json.Marshal(dirs)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.Response.SetBody(dirJson)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (dh *DirectoriesHandler) RegisterRoutes(r *router.Router) {
	group := r.Group("/dirs")
	group.OPTIONS("/{dirId}", func(ctx *fasthttp.RequestCtx) {
        ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
        ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.GET("/{dirId}", middleware.CommonMW(dh.Get))
}

func NewDirsHandler(service DirectoriesService) *DirectoriesHandler {
	return &DirectoriesHandler{
		service: service,
	}
}
