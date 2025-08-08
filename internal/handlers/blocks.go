package handlers

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type BlocksService interface {
	Add() error
	Move() error
	Delete() error
}

type BlocksHandler struct {
	service BlocksService
}

func NewBlocksHandler(service BlocksService) *BlocksHandler {
	return &BlocksHandler{
		service: service,
	}
}

// Add adds new block of code to file
// @Summary Add block
// @Description Adds block of code
// @Tags blocks
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /blocks [post]
func (bh *BlocksHandler) Add(ctx *fasthttp.RequestCtx) {
	err := bh.service.Add()

	if err != nil {
		ctx.Response.SetBody([]byte(err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

// Add adds new block of code to file
// @Summary Add block
// @Description Adds block of code
// @Tags blocks
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /blocks/{id} [put]
func (bh *BlocksHandler) Update(ctx *fasthttp.RequestCtx) {
	err := bh.service.Move()

	if err != nil {
		ctx.Response.SetBody([]byte(err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

// Add adds new block of code to file
// @Summary Add block
// @Description Adds block of code
// @Tags blocks
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /blocks/{id} [delete]
func (bh *BlocksHandler) Delete(ctx *fasthttp.RequestCtx) {
	err := bh.service.Delete()

	if err != nil {
		ctx.Response.SetBody([]byte(err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (bh *BlocksHandler) Compile(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (bh *BlocksHandler) Run(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (bh *BlocksHandler) RegisterRoutes(r *router.Router) {
	group := r.Group("/block")
	group.POST("", bh.Add)
	group.PUT("", bh.Update)
	group.DELETE("", bh.Delete)
}
