package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/dnonakolesax/noted-notes/internal/middleware"
	"github.com/dnonakolesax/noted-notes/internal/model"
	"github.com/google/uuid"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)


type FilesService interface {
	Get(fileID uuid.UUID, userID uuid.UUID) (model.FileDTO, error)
	Remove() error
}

type FileHandler struct {
	service FilesService
}

func NewFileHandler(service FilesService) *FileHandler {
	return &FileHandler{
		service: service,
	}
}

func (fh *FileHandler) Get(ctx *fasthttp.RequestCtx) {
	fileId := ctx.UserValue("fileId").(string)
	//userId := ctx.UserValue("userId").(string)
	
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

	file, err := fh.service.Get(fileUUID, userUUID)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	fileJSON, err := json.Marshal(file)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetBody(fileJSON)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (fh *FileHandler) RegisterRoutes(r *router.Router) {
	group := r.Group("/files")
	group.OPTIONS("/{fileId}", func(ctx *fasthttp.RequestCtx) {
        ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
        ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.GET("/{fileId}", middleware.CommonMW(fh.Get))
}
