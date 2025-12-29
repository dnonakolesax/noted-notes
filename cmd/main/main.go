package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	dbsql "github.com/dnonakolesax/noted-notes/internal/db/sql"
	"github.com/dnonakolesax/noted-notes/internal/handlers/ws"
	"github.com/dnonakolesax/noted-notes/internal/middleware"
	"github.com/dnonakolesax/noted-notes/internal/s3"
	pb "github.com/dnonakolesax/noted-notes/internal/services/auth/proto"
	"github.com/fasthttp/router"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/dnonakolesax/noted-notes/internal/handlers"
	"github.com/dnonakolesax/noted-notes/internal/repos"
	"github.com/dnonakolesax/noted-notes/internal/services"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		slog.Warn("couldn't load .env")
	}

	slog.Info("creating s3")
	s3worker, err := s3.NewS3Worker(os.Getenv("S3_ADDR"))
	slog.Info("created s3")

	if err != nil {
		panic(err)
	}
	dbConfig := dbsql.RDBConfig{
		Address:  os.Getenv("DB_ADDRESS"),
		Port:     5432,
		DBName:   os.Getenv("DB_NAME"),
		Login:    os.Getenv("DB_LOGIN"),
		Password: os.Getenv("DB_PASSWORD"),
	}
	slog.Info("creating pgxconn")
	dbConn, err := dbsql.NewPGXConn(dbConfig)
	if err != nil {
		panic(err)
	}
	slog.Info("created pgxconn")
	slog.Info("creating pgxworker")
	dbWorker, err := dbsql.NewPGXWorker(dbConn)
	if err != nil {
		panic(err)
	}
	slog.Info("created pgxworker")

	conn, err := grpc.NewClient("auth:8801", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewAuthServiceClient(conn)

	blockRepo := repos.NewBlockRepo(*s3worker, *dbWorker)
	fileRepo := repos.NewFilesRepo(*dbWorker)
	dirsRepo := repos.NewDirsRepo(*dbWorker)
	treeRepo := repos.NewFileTreeRepo(dbWorker)
	accessRepo := repos.NewAccessRepo(dbWorker)
	csrfRepo := repos.NewCSRF()

	fileService := services.NewFilesService(fileRepo, blockRepo)
	dirsService := services.NewDirsService(dirsRepo)
	treeService := services.NewTreeService(treeRepo)
	blockService := services.NewBlockService(blockRepo)
	accessService := services.NewAccessService(accessRepo)

	accessMW := middleware.NewAccessMW(accessService)
	authMW := middleware.NewAuthMW(c, slog.Default(), csrfRepo)
	csrfMW := middleware.NewCSRFMW(csrfRepo)

	fileHandler := handlers.NewFileHandler(fileService, accessMW)
	dirsHandler := handlers.NewDirsHandler(dirsService, accessMW)
	treeHandler := handlers.NewFileTreeHandler(treeService, accessMW, accessService)
	blockHandler := handlers.NewBlocksHandler(blockService, accessMW)
	hotDir := "/noted/codes/kernels"

	mgr := ws.NewManager(blockRepo, hotDir)
	socketHandler := ws.NewHandler(mgr, accessService)

	r := router.New()
	rtr := r.Group("/api/v1/fm")
	fileHandler.RegisterRoutes(rtr)
	dirsHandler.RegisterRoutes(rtr)
	socketHandler.RegisterRoutes(rtr)
	treeHandler.RegisterRoutes(rtr)
	blockHandler.RegisterRoutes(rtr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	srv := fasthttp.Server{
		Handler: middleware.CommonMW(authMW.AuthMiddleware(csrfMW.MW(r.Handler))),
	}
	slog.Info("starting server on", slog.String("addr", "127.0.0.1:"+os.Getenv("APP_PORT")))
	go func() {
		err := srv.ListenAndServe(":" + os.Getenv("APP_PORT"))
		if err != nil {
			fmt.Printf("listen and serve returned err: %s \n", err)
		}
	}()
	slog.Info("waiting for signal")
	sig := <-quit
	fmt.Printf("stopped : %s \n", sig.String())
	err = srv.Shutdown()
	if err != nil {
		fmt.Printf("shutdown returned err: %s \n", err)
	}
}
