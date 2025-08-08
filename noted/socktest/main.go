package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"io"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/jackc/pgx/v4/pgxpool"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/valyala/fasthttp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/aws/aws-sdk-go-v2/aws"
  	"github.com/aws/aws-sdk-go-v2/config"
  	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(ctx *fasthttp.RequestCtx) bool { return true },
}

type update struct {
	What    string `json:"type"`
	Message string `json:"message"`
	Offset  int64  `json:"offset"`
	RLength int64  `json:"rlength"`
}

type ConnMessage struct {
	What    string `json:"type"`
	Who string `json:"who"`
	Offset int64 `json:"offset"`
}

type Code struct {
	Text string `json:"text"`
}

type Request struct {
	Body string `json:"body"`
}

type CodeRequest struct {
	Filename string `json:"filename"`
}

type codeFile struct {
	Name string `json:"filename"`
	Data string `json:"code"`
}

type codeResponse struct {
	Data string `json:"body"`
	Error string `json:"error"`
}

var subscribers = make(map[string][]*websocket.Conn)

func getCode(ctx *fasthttp.RequestCtx) {
	var req Request
	_ = json.Unmarshal(ctx.Request.Body(), &req)
	// fmt.Println(ctx.Request.Body())
	// _ = json.Unmarshal(req.Body, &codeReq)

	fmt.Println(req.Body)

	bytes, _ := os.ReadFile(req.Body)
	code := Code{Text: string(bytes)}
	resp, _ := json.Marshal(code)
	_, _ = ctx.Write(resp)
	ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
}

type DirsRequest struct {
	Parent string `json:"parent"`
}

type DirsResponse struct {
	FileId string `json:"fileId"`
	Name string `json:"name"`
	Dir bool `json:"dir"`
}

func getDirs(ctx *fasthttp.RequestCtx, db *pgxpool.Pool) {
	id := "0416603d-9a5c-4290-a1dd-62babfea991e"
	var req DirsRequest
	_ = json.Unmarshal(ctx.Request.Body(), &req)
	dbReqStr := "SELECT file_id, filename, dir FROM files JOIN file_access USING (file_id) WHERE (public = true OR user_id = $1) AND parent_id = $2"
	resp, err := db.Query(context.Background(), dbReqStr, id, req.Parent)

	if err != nil {
		fmt.Println(err.Error())
		ctx.Response.SetStatusCode(500)
		return
	}

	uzbek := DirsResponse{}
	uzbeks := make([]DirsResponse, 0)
	defer resp.Close()
	for resp.Next() {
		err = resp.Scan(&uzbek.FileId,&uzbek.Name, &uzbek.Dir)
		if err != nil {
			fmt.Println(err.Error())
			ctx.Response.SetStatusCode(500)
			return
		}
		uzbeks = append(uzbeks, uzbek)
	}

	htmlResp, err := json.Marshal(uzbeks)
	if err != nil {
		ctx.Response.SetStatusCode(500)
		return
	}

	ctx.Response.SetBody(htmlResp)
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

}

func compile (ctx *fasthttp.RequestCtx, ch *amqp.Channel, q amqp.Queue) {
	err := ch.PublishWithContext(context.Background(),
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:       ctx.Request.Body(),
			})

	if err != nil {
		log.Fatalf("failed to publish a message. Error: %s", err)
		ctx.Response.SetStatusCode(400)
		return
	}

	var code codeFile
	var dat []byte

	_ = json.Unmarshal(ctx.Request.Body(), &code)

	time.Sleep(2*time.Second)

	dat, _ = os.ReadFile("/tmp/kernel228/" + code.Name + ".txt")
	fmt.Println("/tmp/kernel228/" + code.Name + ".txt")
	

	codeResp := codeResponse{Data: string(dat), Error: ""}
	jsonData, _ := json.Marshal(codeResp)
	_, _ = ctx.Write(jsonData)

	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
}

func SockNigger(ctx *fasthttp.RequestCtx) {
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		connMesage := ConnMessage{What: "conn", Who: "domnakolesax", Offset: 0}
		connMsg, _ := json.Marshal(connMesage)
		for _, sub := range(subscribers["uzbek"]) {
			_ = sub.WriteMessage(1, connMsg)
		}
		subscribers["uzbek"] = append(subscribers["uzbek"], conn)
		//fmt.Printf("tojik")
		for {
			messageType, p, err := conn.ReadMessage()
			fmt.Printf("Address of pointer: %p\n", &conn)
			if messageType == -1 {
				fmt.Printf("closed")
			}
			if err != nil {
				fmt.Printf("closed2")
				log.Println(err)
				log.Println(messageType)
				return
			}
			fmt.Printf("%d", messageType)
			var descriptor, _ = os.OpenFile("uzbek.txt", os.O_WRONLY, 0666)
			dat, _ := os.ReadFile("uzbek.txt")
			var upd update
			_ = json.Unmarshal(p, &upd)
			if upd.Message == "" {
				fmt.Printf("no message!")
				_ = descriptor.Truncate(upd.Offset)
				_, _ = descriptor.WriteAt(dat[upd.Offset+upd.RLength:], upd.Offset)
			} else {
				_, _ = descriptor.Seek(upd.Offset, 0)
				_, _ = descriptor.Write([]byte(upd.Message))
				_, _ = descriptor.Write(dat[upd.Offset:])
			}
			updJSON, _ := json.Marshal(upd)
			for _, sub := range(subscribers["uzbek"]) {
				if sub == conn {
					continue
				}
				_ = sub.WriteMessage(1, updJSON)
			}
			descriptor.Close()
		}
	})

	if err != nil {
		log.Println(err)
		return
	}
}

type blocksRequest struct {
	FileId string `json:"parent"`
}

type Block struct {
	Language string `json:"language"`
	Code string `json:"code"`
}

type blocksResponse struct {
	Blocks []Block `json:"blocks"`
}

type codeMetadata struct {
	Id string `json:"id"`
	LastUpdated string `json:"last"`
	Owner string `json:"owner"`
	Blocks []string `json:"blocks"`
}

func getBlocks (ctx *fasthttp.RequestCtx, collection *mongo.Collection, bb *BucketBasics) {
	var req blocksRequest
	_ = json.Unmarshal(ctx.Request.Body(), &req)

	filter := bson.M{"id": req.FileId}
	var result codeMetadata

	err := collection.FindOne(context.Background(), filter).Decode(&result)

	if err != nil {
		fmt.Println(err)
		ctx.Response.SetStatusCode(400)
		return
	}

	var block Block
	respBlocks := make([]Block, 0)
	for i, cblock := range(result.Blocks) {
		if cblock != "" {
			dblock, _ := bb.DownloadFile(context.Background(), "noted", cblock)
			if i == 1 {
				block = Block{Language: "md", Code: string(dblock)}
			} else {
				block = Block{Language: "go", Code: string(dblock)}
			}
			respBlocks = append(respBlocks, block)
		}
	}

	finalResponse := blocksResponse{respBlocks}

	jsonBlocks, _ := json.Marshal(finalResponse)
	ctx.Response.SetBody(jsonBlocks)
	ctx.Response.SetStatusCode(200)
	ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
	ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
}






type BucketBasics struct {
	S3Client *s3.Client
  }  

func (basics BucketBasics) DownloadFile(ctx context.Context, bucketName string, objectKey string) ([]byte, error) {
	result, err := basics.S3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		log.Printf("Couldn't get object %v:%v. Here's why: %v\n", bucketName, objectKey, err)
		return []byte{}, err
	}
	defer result.Body.Close()
	//file, err := os.Create(fileName)
	// if err != nil {
	// 	log.Printf("Couldn't create file %v. Here's why: %v\n", fileName, err)
	// 	return err
	// }
	// defer file.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		log.Printf("Couldn't read object body from %v. Here's why: %v\n", objectKey, err)
		return []byte{}, err
	}
	return body, nil
}



func main() {

	cfg, err := config.LoadDefaultConfig(context.TODO())

	addr := "https://storage.yandexcloud.net"

	cfg.BaseEndpoint = &addr

	if err != nil {
		log.Fatal(err)
	}

	// Создаем клиента для доступа к хранилищу S3
	s3client := s3.NewFromConfig(cfg)
	bucketBasics := BucketBasics{S3Client: s3client}





	descriptor, er := os.OpenFile("uzbek228.txt", os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if er != nil {
		fmt.Printf("%s", er.Error())
	}
	defer descriptor.Close()
	_, er = descriptor.Write([]byte("uzbek228"))

	if er != nil {
		fmt.Printf("%s", er.Error())
	}


	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://172.27.0.2:27017"))
	if err != nil {
		log.Fatal(err)
	}
	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	collection := client.Database("noted").Collection("metadatas")


	conn, err := amqp.Dial("amqp://guest:guest@172.26.0.2:5672/")
	if err != nil {
		log.Fatalf("unable to open connect to RabbitMQ server. Error: %s", err)
	}

	defer func() {
		_ = conn.Close() // Закрываем подключение в случае удачной попытки подключения
	}()
	
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel. Error: %s", err)
	}

	defer func() {
		_ = ch.Close() // Закрываем подключение в случае удачной попытки подключения
	}()

	q, err := ch.QueueDeclare(
		"prepare", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare a queue. Error: %s", err)
	}


	db, err := pgxpool.Connect(context.Background(), fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable",
		"kopilka",
		"12345",
		"172.22.0.2",
		5432,
		"noted"))
	if err != nil {
		fmt.Printf("fail open postgres")
	}
	defer db.Close()


	m := func(ctx *fasthttp.RequestCtx) {
		if string(ctx.Request.Header.Method()) == "OPTIONS" {
			fmt.Printf("OPTIONS")
			ctx.Response.Header.Add("Access-Control-Allow-Origin", "*")
			ctx.Response.Header.Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			ctx.Response.Header.Add("Access-Control-Allow-Headers", "*")
		} else {
			switch string(ctx.Path()) {
			case "/nigger":
				SockNigger(ctx)
			case "/getcode":
				getCode(ctx)
			case "/uzbek":
				compile(ctx, ch, q)	
			case "/dirs":
				getDirs(ctx, db)	
			case "/blocks":
				getBlocks(ctx, collection, &bucketBasics)	
			default:
				ctx.Error("not found", fasthttp.StatusNotFound)
			}
		}
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	srv := &fasthttp.Server{
		Handler: m,
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
