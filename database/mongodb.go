package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func RequireUniqueEmailFields(collection *mongo.Collection) error {
	cur, err := collection.Indexes().List(context.TODO())
	if err != nil {
		return err
	}

	idxExists := cur.Next(context.TODO())
	if !idxExists {
		collection.Indexes().CreateOne(
			context.TODO(),
			mongo.IndexModel{
				Keys:    bson.D{{"email", 1}},
				Options: options.Index().SetUnique(true),
			},
		)
	}

	return nil
}
