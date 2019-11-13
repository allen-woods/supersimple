package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// GoMongo provides connectivity to a MongoDB database.
func GoMongo(db string, col string) (context.Context, *mongo.Collection) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		log.Fatal("No database was found. Is MongoDB running?")
	}

	collection := client.Database(db).Collection(col)

	return ctx, collection
}

func RequireUniqueEmailFields(ctx context.Context, collection *mongo.Collection) error {
	cur, err := collection.Indexes().List(ctx)
	if err != nil {
		return err
	}

	idxExists := cur.Next(ctx)
	if !idxExists {
		collection.Indexes().CreateOne(
			ctx,
			mongo.IndexModel{
				Keys:    bson.D{{"email", 1}},
				Options: options.Index().SetUnique(true),
			},
		)
	}

	return nil
}
