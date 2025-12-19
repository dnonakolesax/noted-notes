package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	dbsql "github.com/dnonakolesax/noted-notes/internal/db/sql"
	"github.com/dnonakolesax/noted-notes/internal/handlers/ws"
	"github.com/dnonakolesax/noted-notes/internal/s3"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-notes/internal/handlers"
	"github.com/dnonakolesax/noted-notes/internal/repos"
	"github.com/dnonakolesax/noted-notes/internal/services"
)

func main() {
	s3worker, err := s3.NewS3Worker("https://storage.yandexcloud.net")

	if err != nil {
		panic(err)
	}
	dbConfig := dbsql.RDBConfig{
		Address: "gobddocker-postgres-1",
		Port: 5432,
		DBName: "noted",
		Login: "kopilka",
		Password: "12345",
	}
	dbConn, err := dbsql.NewPGXConn(dbConfig)
	if err != nil {
		panic(err)
	}
	dbWorker, err := dbsql.NewPGXWorker(dbConn)
	if err != nil {
		panic(err)
	}

	blockRepo := repos.NewBlockRepo(*s3worker, *dbWorker)
	fileRepo := repos.NewFilesRepo(*dbWorker)
	dirsRepo := repos.NewDirsRepo(*dbWorker)
	treeRepo := repos.NewFileTreeRepo(dbWorker)

	fileService := services.NewFilesService(fileRepo, blockRepo)
	dirsService := services.NewDirsService(dirsRepo)
	treeService := services.NewTreeService(treeRepo)
	blockService := services.NewBlockService(blockRepo)

	fileHandler := handlers.NewFileHandler(fileService)
	dirsHandler := handlers.NewDirsHandler(dirsService)
	treeHandler := handlers.NewFileTreeHandler(treeService)
	blockHandler := handlers.NewBlocksHandler(blockService)
	hotDir := "/noted/codes/kernels"

	mgr := ws.NewManager(blockRepo, hotDir)
	socketHandler := ws.NewHandler(mgr)

	rtr := router.New()
	fileHandler.RegisterRoutes(rtr)
	dirsHandler.RegisterRoutes(rtr)
	socketHandler.RegisterRoutes(rtr)
	treeHandler.RegisterRoutes(rtr)
	blockHandler.RegisterRoutes(rtr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	srv := fasthttp.Server{
		Handler: rtr.Handler,
	}
	go func() {
		err := srv.ListenAndServe("0.0.0.0:5004")
		if err != nil {
			fmt.Printf("listen and serve returned err: %s \n", err)
		}
	}()

	sig := <-quit
	fmt.Printf("stopped : %s \n", sig.String())
	err = srv.Shutdown()
	if err != nil {
		fmt.Printf("shutdown returned err: %s \n", err)
	}
}
