package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/dnonakolesax/noted-notes/internal/consts"
	"github.com/dnonakolesax/noted-notes/internal/middleware"
	"github.com/dnonakolesax/noted-notes/internal/model"
	"github.com/google/uuid"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type FilesService interface {
	Get(fileID uuid.UUID, userID uuid.UUID) (model.FileDTO, error)
	Delete(fileID uuid.UUID) error
}

type FileHandler struct {
	service  FilesService
	accessMW *middleware.AccessMW
}

func NewFileHandler(service FilesService, accessMW *middleware.AccessMW) *FileHandler {
	return &FileHandler{
		service:  service,
		accessMW: accessMW,
	}
}

func (fh *FileHandler) Get(ctx *fasthttp.RequestCtx) {
	fileId := ctx.UserValue("fileID").(string)
	userId := ctx.UserValue(consts.CtxUserIDKey).(string)

	//userId := "0416603d-9a5c-4290-a1dd-62babfea991e"
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
	file.Rights = ctx.UserValue("access").(string)

	fileJSON, err := json.Marshal(file)

	if err != nil {
		fmt.Printf("Error: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetBody(fileJSON)
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (fh *FileHandler) Delete(ctx *fasthttp.RequestCtx) {
	fileId := ctx.UserValue("fileID").(string)

	fileUUID, err := uuid.Parse(fileId)

	if err != nil {
		fmt.Printf("Error uuid parse: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	err = fh.service.Delete(fileUUID)
	if err != nil {
		fmt.Printf("Error uuid parse: %v at %s\n", err, ctx.URI().Path())
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (fh *FileHandler) RegisterRoutes(r *router.Group) {
	group := r.Group("/files")
	group.OPTIONS("/{fileID}", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Access-Control-Allow-Origin", consts.URL + "/files/*")
		ctx.Response.Header.Add("Access-Control-Allow-Methods", "GET, DELETE")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.GET("/{fileID}", fh.accessMW.Read(fh.Get))
	group.DELETE("/{fileID}", fh.accessMW.Own(fh.Delete))
}
