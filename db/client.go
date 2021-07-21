package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbClient struct {
	MongoClient *mongo.Client
	Ctx         context.Context
	Collection  *mongo.Collection
}

func NewDBClient(ctx context.Context) (cli *DbClient) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	cli = &DbClient{
		MongoClient: client,
		Ctx:         ctx,
	}

	return cli
}

func (c *DbClient) SetCollection(db string, collection string) {
	c.Collection = c.MongoClient.Database(db).Collection(collection)
}

func (c *DbClient) GetOne(filter primitive.D, decode interface{}) error {
	err := c.Collection.FindOne(c.Ctx, filter).Decode(&decode)
	return err
}
