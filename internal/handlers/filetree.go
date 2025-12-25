package handlers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dnonakolesax/noted-notes/internal/consts"
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
	authMW        *middleware.AuthMW
}

func NewFileTreeHandler(filesService FileTreeService, accessMW *middleware.AccessMW, accessService AccessService, authMW *middleware.AuthMW) *FileTreeHandler {
	return &FileTreeHandler{
		ftreeService:  filesService,
		accessMW:      accessMW,
		accessService: accessService,
		authMW:        authMW,
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

	right, err := fh.accessService.Get(parentUUIDstr, ctx.UserValue(consts.CtxUserIDKey).(string), false)

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
		ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.PUT("/{dirID}", middleware.CommonMW(fh.authMW.AuthMiddleware(fh.accessMW.Write(fh.Add))))

	group.OPTIONS("/rename/{fileID}", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.PATCH("/rename/{fileID}", middleware.CommonMW(fh.authMW.AuthMiddleware(fh.accessMW.Own(fh.Rename))))

	group.OPTIONS("/move/{fileID}", func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.PATCH("/move/{fileID}", middleware.CommonMW(fh.authMW.AuthMiddleware(fh.accessMW.Own(fh.Move))))
}
