package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CatFactWorker struct {
	client *mongo.Client
}

type Server struct {
	client *mongo.Client
}

func NewServer(c *mongo.Client) *Server {
	return &Server{
		client: c,
	}
}

func (s *Server) handlerGetAllFacts(c *fiber.Ctx) error {
	coll := s.client.Database("catfact").Collection("facts")

	query := bson.M{}
	cursor, err := coll.Find(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}
	results := []bson.M{}
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}
	return c.JSON(results)
}

func NewCatFactWorker(c *mongo.Client) *CatFactWorker {
	return &CatFactWorker{client: c}
}

func (cfw *CatFactWorker) start() error {
	coll := cfw.client.Database("catfact").Collection("facts")
	ticker := time.NewTicker(2 * time.Second)

	for {
		resp, err := http.Get("https://catfact.ninja/fact")
		if err != nil {
			return err
		}
		var catFact bson.M // it is a map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&catFact); err != nil {
			return err
		}
		e, err := coll.InsertOne(context.TODO(), catFact)
		_ = e
		if err != nil {
			return err
		}
		fmt.Println(catFact)
		<-ticker.C
	}
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(client)
	worker := NewCatFactWorker(client)
	// go worker.start()
	server := NewServer(worker.client)
	app := fiber.New()
	app.Get("/facts", server.handlerGetAllFacts)
	app.Listen(":8000")
}
