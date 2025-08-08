package main

import (
	"context"
	"log"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type codeMetadata struct {
	Id string `json:"id"`
	LastUpdated string `json:"last"`
	Owner string `json:"owner"`
	Blocks []string `json:"blocks"`
}

func main () {
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

	codeBlocks := make([]string, 1)
	codeBlocks = append(codeBlocks, "block_2de5g6h72de5g6h7")
	codeBlocks = append(codeBlocks, "block_2de5g6h72de5g6h8")
	codeBlocks = append(codeBlocks, "block_2de5g6h72de5g6h9")
	uzbek := codeMetadata{Id: "53cbef31-1891-4969-ad26-33947b6ebfb2",LastUpdated: "2024-10-29",Owner: "domnakolesax",Blocks: codeBlocks}

	insertResult, err := collection.InsertOne(context.TODO(), uzbek)

	fmt.Println(insertResult.InsertedID)
	if err != nil {
		log.Fatal(err)
	}

	filter := bson.M{"id": "53cbef31-1891-4969-ad26-33947b6ebfb2"}
	var result codeMetadata

	err = collection.FindOne(context.Background(), filter).Decode(&result)
	
	fmt.Println(result.Id)
	fmt.Println(result.LastUpdated)
	fmt.Println(result.Owner)
	for i, val := range(result.Blocks) {
		if val != "" {
			fmt.Println(i)
			fmt.Println(val)
		}
	}

	if err != nil {
		log.Fatal(err)
	}
}
