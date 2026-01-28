package config

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	defaultCollection string
	Database          *mongo.Database
}

func NewMongoClient() *MongoClient {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	var client *mongo.Client
	mongoUri := os.Getenv("MONGO_URI")
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatal(err)
	}
	return &MongoClient{
		defaultCollection: "configurations",
		Database:          client.Database("configurations"),
	}
}

func (mc *MongoClient) ExecuteSelectionQuery(query bson.M, coll ...string) ([]bson.M, error) {
	collection := mc.defaultCollection
	if len(coll) > 0 {
		collection = coll[0]
	}
	var opts *options.FindOptions
	opts = options.Find()
	var cur *mongo.Cursor
	cur, err := mc.Database.Collection(collection).Find(
		context.TODO(),
		query,
		opts,
	)
	if err != nil {
		// ErrNoDocuments means that the filter did not match any documents in
		// the collection.
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		log.Panic(err)
	}
	var result = make([]bson.M, 0)
	for cur.Next(context.TODO()) {
		var elem bson.M
		err := cur.Decode(&elem)
		if err != nil {
			log.Panic(err)
		}
		result = append(result, elem)
	}

	return result, nil
}

func (mc *MongoClient) InsertData(data bson.M, coll ...string) (InsertedID interface{}, err error) {
	collection := mc.defaultCollection
	if len(coll) > 0 {
		collection = coll[0]
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	result, err := mc.Database.Collection(collection).InsertOne(ctx, data)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return result.InsertedID, nil
}

func (mc *MongoClient) UpdateData(filter bson.M, update bson.M, coll ...string) (ModifiedCount int64, err error) {
	collection := mc.defaultCollection
	if len(coll) > 0 {
		collection = coll[0]
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	result, err := mc.Database.Collection(collection).ReplaceOne(ctx, filter, update)
	if err != nil {
		log.Fatal(err)
		return 0, err
	}
	log.Println("Modified count:", result.ModifiedCount)
	return result.ModifiedCount, nil
}
