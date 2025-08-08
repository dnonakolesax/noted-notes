package handlers

import (
	"github.com/fasthttp/router"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
)

type FileTreeService interface {
	Add(userId uuid.UUID, name string, isDir bool, isPublic bool) (uuid.UUID, error)
	Rename(userId uuid.UUID,id uuid.UUID, newName string) error
	Move(userId uuid.UUID,id uuid.UUID, newParentId uuid.UUID) error
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
	
}

func (fh *FileTreeHandler) Rename(ctx *fasthttp.RequestCtx) {

}

func (fh *FileTreeHandler) Move(ctx *fasthttp.RequestCtx) {

}

func (fh *FileTreeHandler) ChangePrivacy(ctx *fasthttp.RequestCtx) {

}

func (fh *FileTreeHandler) GrantAccess(ctx *fasthttp.RequestCtx) {

}

func (fh *FileTreeHandler) RegisterRoutes(r *router.Router) {

}