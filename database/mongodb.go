package database

import (
	context
	log

	"go.mongodb.org/mongo-driver/mongo"
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