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

type FileTreeService interface {
	Add(filename string, uuid uuid.UUID, isDir bool, parentDir uuid.UUID) error
	Rename(id uuid.UUID, newName string) error
	Move(fileUUID uuid.UUID, newParent uuid.UUID) error

	ChangePrivacy(userId uuid.UUID,id uuid.UUID, isPublic bool) error
	GrantAccess(userId uuid.UUID, id uuid.UUID, targetUserId uuid.UUID, accessType string) error
}

type FileTreeHandler struct {
	ftreeService FileTreeService
}

func NewFileTreeHandler(filesService FileTreeService) *FileTreeHandler {
	return &FileTreeHandler{
		ftreeService: filesService,
	}
}

func (fh *FileTreeHandler) Add(ctx *fasthttp.RequestCtx) {
	var dto model.FileInTreeDTO

	err := json.Unmarshal(ctx.PostBody(), &dto)

	if err != nil {
		fmt.Printf("error unmarshal json: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	fileUUID, err := uuid.NewV7()

	if err != nil {
		fmt.Printf("error parse file uuid: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	parentUUID, err := uuid.Parse(dto.ParentDir)

	if err != nil {
		fmt.Printf("error parse parent uuid: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	err = fh.ftreeService.Add(dto.Name, fileUUID, dto.IsDir, parentUUID)

	if err != nil {
		fmt.Printf("error treeservice add: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	returnDTO := model.FileInTreeRDTO{ID: fileUUID.String()}

	bts, err := json.Marshal(returnDTO)

	if err != nil {
		fmt.Printf("error tree add marshal: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.Response.SetBody(bts)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (fh *FileTreeHandler) Rename(ctx *fasthttp.RequestCtx) {
	fileId := ctx.UserValue("fileId").(string)
	var dto struct {
		Name string `json:"name,omitempty"`
	}

	err := json.Unmarshal(ctx.PostBody(), &dto)

	if err != nil {
		fmt.Printf("error unmarshal json: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	fileUUID, err := uuid.Parse(fileId)

	if err != nil {
		fmt.Printf("error parse file uuid: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	err = fh.ftreeService.Rename(fileUUID, dto.Name)
	
	if err != nil {
		fmt.Printf("error treeservice rename: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (fh *FileTreeHandler) Move(ctx *fasthttp.RequestCtx) {
	fileId := ctx.UserValue("fileId").(string)
	var dto struct {
		ParentID string `json:"parent_dir,omitempty"`
	}

	err := json.Unmarshal(ctx.PostBody(), &dto)

	if err != nil {
		fmt.Printf("error unmarshal json: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	fileUUID, err := uuid.Parse(fileId)

	if err != nil {
		fmt.Printf("error parse file uuid: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	parentUUID, err := uuid.Parse(dto.ParentID)

	if err != nil {
		fmt.Printf("error parse parent uuid: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	err = fh.ftreeService.Move(fileUUID, parentUUID)

	if err != nil {
		fmt.Printf("error treeservice move: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (fh *FileTreeHandler) ChangePrivacy(ctx *fasthttp.RequestCtx) {

}

func (fh *FileTreeHandler) GrantAccess(ctx *fasthttp.RequestCtx) {

}

func (fh *FileTreeHandler) RegisterRoutes(r *router.Router) {
	group := r.Group("/tree")
	group.OPTIONS("/", func(ctx *fasthttp.RequestCtx) {
        ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
        ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.PUT("/", middleware.CommonMW(fh.Add))
	
	group.OPTIONS("/rename/{fileId}", func(ctx *fasthttp.RequestCtx) {
        ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
        ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.PATCH("/rename/{fileId}", middleware.CommonMW(fh.Rename))
	
	group.OPTIONS("/move/{fileId}", func(ctx *fasthttp.RequestCtx) {
        ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
        ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.PATCH("/move/{fileId}", middleware.CommonMW(fh.Move))
}
