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

type BlocksService interface {
	Save(block model.BlockVO) error

	Delete(id string) error

	Move(id string, parentID string, direction string) error
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
	var block model.BlockVO

	err := json.Unmarshal(ctx.Request.Body(), &block)
	if err != nil {
		fmt.Printf("error unmarshal: %v", err)
		ctx.Response.SetBody([]byte(err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	newID, err := uuid.NewV7()

	if err != nil {
		fmt.Printf("error creating uuid: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	block.ID = newID.String()
	err = bh.service.Save(block)

	if err != nil {
		fmt.Printf("error save: %v", err)
		ctx.Response.SetBody([]byte(err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	dto := model.NewBlockRDTO{ID: block.ID}

	bts, err := json.Marshal(dto)
	
	if err != nil {
		fmt.Printf("error marshaling block response: %v", err)
		ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.SetBody(bts)
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
	id := ctx.Request.UserValue("id")
	strID, ok := id.(string)

	if !ok {
		ctx.Response.SetBody([]byte("no block id passed"))
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	newParent := ctx.QueryArgs().Peek("parent_id")

	if newParent == nil {
		ctx.Response.SetBody([]byte("no parent id passed"))
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	_, err := uuid.Parse(string(newParent))

	if err != nil {
		ctx.Response.SetBody([]byte("bad parent id passed"))
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	direction := ctx.QueryArgs().Peek("dir")

	if direction == nil {
		ctx.Response.SetBody([]byte("bad direction passed"))
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	err = bh.service.Move(strID, string(newParent), string(direction))

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
	id := ctx.Request.UserValue("id")
	strID, ok := id.(string)

	if !ok {
		ctx.Response.SetBody([]byte("no block id passed"))
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	err := bh.service.Delete(strID)

	if err != nil {
		ctx.Response.SetBody([]byte(err.Error()))
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.Response.SetStatusCode(fasthttp.StatusOK)
}

func (bh *BlocksHandler) RegisterRoutes(r *router.Router) {
	group := r.Group("/block")

	group.OPTIONS("/", func(ctx *fasthttp.RequestCtx) {
        ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
        ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.POST("/", middleware.CommonMW(bh.Add))

	group.OPTIONS("/{id}", func(ctx *fasthttp.RequestCtx) {
        ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
        ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT, PATCH")
		ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
	})
	group.PATCH("/{id}", middleware.CommonMW(bh.Update))
	group.DELETE("/{id}", middleware.CommonMW(bh.Delete))
}
