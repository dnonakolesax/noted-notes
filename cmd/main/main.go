package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	dbsql "github.com/dnonakolesax/noted-notes/internal/db/sql"
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
		Address: "172.20.0.2",
		Port: 5432,
		DBName: "noted",
		Login: "dnonakolesax",
		Password: "228",
	}
	dbConn, err := dbsql.NewPGXConn(dbConfig)
	if err != nil {
		panic(err)
	}
	dbWorker, err := dbsql.NewPGXWorker(dbConn)
	if err != nil {
		panic(err)
	}

	blockRepo := repos.NewBlockRepo(*s3worker)
	fileRepo := repos.NewFilesRepo(*dbWorker)
	dirsRepo := repos.NewDirsRepo(*dbWorker)

	fileService := services.NewFilesService(fileRepo, blockRepo)
	dirsService := services.NewDirsService(dirsRepo)
	socketService := services.NewSocketService()

	fileHandler := handlers.NewFileHandler(fileService)
	dirsHandler := handlers.NewDirsHandler(dirsService)
	socketHandler := handlers.NewSocketHandler(socketService)

	rtr := router.New()
	fileHandler.RegisterRoutes(rtr)
	dirsHandler.RegisterRoutes(rtr)
	socketHandler.RegisterRoutes(rtr)

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