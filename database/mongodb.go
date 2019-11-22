package database

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoDBObject struct {
	c   interface{}
	ctx interface{}
}

func CreateClient() error {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return err
	}
	mongoDBObject.c = client
	return nil
}

func AssignContext(ctx context.Context) (context.Context, error) {
	mongoDBObject.ctx = ctx
	mdbCtx, ok := mongoDBObject.ctx.(context.Context)
	if !ok {
		err := errors.New("Unable to assign context to database package!")
		return nil, err
	}

	return mdbCtx, nil
}

func GetClient() (*mongo.Client, error) {
	client, ok := mongoDBObject.c.(*mongo.Client)
	if !ok {
		err := errors.New("Unable to get client from database package!")
		return nil, err
	}

	return client, nil
}

func GetContext() (context.Context, error) {
	ctx, ok := mongoDBObject.ctx.(context.Context)
	if !ok {
		err := errors.New("Unable to get context from database package!")
		return nil, err
	}

	return ctx, nil
}

func RequireUniqueEmailFields(collection *mongo.Collection) error {
	ctx, err := GetContext()
	if err != nil {
		return err
	}

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
