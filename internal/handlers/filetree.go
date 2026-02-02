package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dnonakolesax/noted-notes/internal/consts"
	"github.com/dnonakolesax/noted-notes/internal/middleware"
	"github.com/dnonakolesax/noted-notes/internal/model"
	"github.com/dnonakolesax/noted-notes/internal/xerrors"
	"github.com/fasthttp/router"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

type FileTreeService interface {
	Add(filename string, uuid uuid.UUID, isDir bool, parentDir uuid.UUID, userID string) error
	Rename(id uuid.UUID, newName string) error
	Move(fileUUID uuid.UUID, newParent uuid.UUID) error

	ChangePrivacy(userId uuid.UUID, id uuid.UUID, isPublic bool) error
	GrantAccess(userId uuid.UUID, id uuid.UUID, targetUserId uuid.UUID, accessType string) error
}

type AccessService interface {
	Get(fileID string, userID string, byBlock bool) (string, error)
}

type FileTreeHandler struct {
	ftreeService  FileTreeService
	accessService AccessService // TODO: отрефакторить, чтобы как-нибудь одно что-нибудб
	accessMW      *middleware.AccessMW
}

func NewFileTreeHandler(filesService FileTreeService, accessMW *middleware.AccessMW, accessService AccessService) *FileTreeHandler {
	return &FileTreeHandler{
		ftreeService:  filesService,
		accessMW:      accessMW,
		accessService: accessService,
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

	parentUUIDstr := ctx.UserValue("dirID").(string)
	parentUUID, err := uuid.Parse(parentUUIDstr)

	if err != nil {
		fmt.Printf("error parse parent uuid: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	
	userID := ctx.Request.UserValue(consts.CtxUserIDKey)
	if userID == nil {
		slog.Warn("userid is nil")
		ctx.Response.SetStatusCode(fasthttp.StatusUnauthorized)
		return 
	}

	right, err := fh.accessService.Get(parentUUIDstr, userID.(string), false)

	if err != nil {
		fmt.Printf("error checking new parent dir rights: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
		return
	}

	if !strings.Contains(right, "w") {
		fmt.Printf("error: new parent dir rights has no w, actually: %s", right)
		ctx.Response.SetStatusCode(fasthttp.StatusForbidden)
		return
	}

	err = fh.ftreeService.Add(dto.Name, fileUUID, dto.IsDir, parentUUID, userID.(string))

	if err != nil {
		if errors.Is(err, xerrors.ErrInvalidFileName) {
			ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
			return
		}
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
	fileId := ctx.UserValue("fileID").(string)
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
		
		if errors.Is(err, xerrors.ErrInvalidFileName) {
			ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
			return
		}
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (fh *FileTreeHandler) Move(ctx *fasthttp.RequestCtx) {
	fileId := ctx.UserValue("fileID").(string)
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

func (fh *FileTreeHandler) RegisterRoutes(r *router.Group) {
	group := r.Group("/tree")
	group.OPTIONS("/", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Access-Control-Allow-Origin", consts.URL + "/files/*")
		ctx.Response.Header.Add("Access-Control-Allow-Methods", "PUT")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.PUT("/{dirID}", fh.accessMW.Write(fh.Add))

	group.OPTIONS("/rename/{fileID}", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Access-Control-Allow-Origin", consts.URL + "/files/*")
		ctx.Response.Header.Add("Access-Control-Allow-Methods", "PATCH")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.PATCH("/rename/{fileID}", fh.accessMW.Own(fh.Rename))

	group.OPTIONS("/move/{fileID}", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Access-Control-Allow-Origin", consts.URL + "/files/*")
		ctx.Response.Header.Add("Access-Control-Allow-Methods", "PATCH")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.PATCH("/move/{fileID}", fh.accessMW.Own(fh.Move))
}
