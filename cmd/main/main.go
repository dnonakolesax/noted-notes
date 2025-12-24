package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	dbsql "github.com/dnonakolesax/noted-notes/internal/db/sql"
	"github.com/dnonakolesax/noted-notes/internal/handlers/ws"
	"github.com/dnonakolesax/noted-notes/internal/middleware"
	"github.com/dnonakolesax/noted-notes/internal/s3"
	"github.com/fasthttp/router"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-notes/internal/handlers"
	"github.com/dnonakolesax/noted-notes/internal/repos"
	"github.com/dnonakolesax/noted-notes/internal/services"
)

func migrate(config dbsql.RDBConfig) {
	db, err := sql.Open("pgx", fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		config.Login,
		config.Password,
		config.Address,
		config.Port,
		config.DBName))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	goose.SetDialect("postgres")

	if err := goose.Up(db, "./db/migrations"); err != nil {
		log.Fatal(err)
	}
}

func main() {
	err := godotenv.Load()

	if err != nil {
		panic(err)
	}

	s3worker, err := s3.NewS3Worker(os.Getenv("S3_ADDR"))

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
	migrate(dbConfig)
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
	accessRepo := repos.NewAccessRepo(dbWorker)

	fileService := services.NewFilesService(fileRepo, blockRepo)
	dirsService := services.NewDirsService(dirsRepo)
	treeService := services.NewTreeService(treeRepo)
	blockService := services.NewBlockService(blockRepo)
	accessService := services.NewAccessService(accessRepo)

	accessMW := middleware.NewAccessMW(accessService)

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
		Handler: r.Handler,
	}
	go func() {
		err := srv.ListenAndServe("127.0.0.1:" + os.Getenv("APP_PORT"))
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
