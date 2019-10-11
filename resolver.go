package supersimple

import (
	"context"
	"log"
	"time"

	supersimple "github.com/allen-woods/supersimple/models"
	"go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

// RESTful Routes
// Index,	Create,	New,	Show,	Update,	Delete,	Edit
// Get,		Post,		Get,	Get,	Patch,	Delete,	Get
//
// MongoDB Collection Methods
// Aggregate (relations)
// BulkWrite
// Clone
// CountDocuments
// Database
// DeleteMany
// DeleteOne
// Distinct
// Drop
// EstimatedDocumentCount
// Find ("Index" route)
// FindOne (returns doc) <-- NEED ("Show" route)
// * FindOneAndDelete (returns doc)
// FindOneAndReplace (put, returns doc) <-- NEED
// * FindOneAndUpdate (patch, returns doc)
// Indexes
// InsertMany
// * InsertOne
// Name
// ReplaceOne
// UpdateMany
// UpdateOne
// Watch (This is for ChangeStream / Subscriptions)

/* What I have:
Create - one
Read - all
Update - one
Delete - one
*/

/* What I need:
Create - many
Read - one
Update - many
Update - all
Delete - many
Delete - all
*/
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

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateUser(ctx context.Context, input supersimple.NewUser) (*supersimple.User, error) {
	ctx, collection := GoMongo("simple", "users")

	u := &supersimple.User{
		Name: input.Name,
	}

	res, err := collection.InsertOne(ctx, *u)
	if err != nil {
		log.Println(err)
	}
	id := res.InsertedID
	u.ID = id.(primitive.ObjectID)
	return u, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, id primitive.ObjectID, name string) (*supersimple.User, error) {
	ctx, collection := GoMongo("simple", "users")

	filter := bson.D{
		{"_id", id},
	}

	update := bson.D{
		{
			"$set", bson.D{
				{"name", name},
			},
		},
	}

	opts := options.FindOneAndUpdate()
	opts.SetReturnDocument(options.After)

	var u supersimple.User

	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&u)
	if err != nil {
		log.Println("Error:", err)
	}
	return &u, nil
}

func (r *mutationResolver) DeleteUser(ctx context.Context, id primitive.ObjectID) (*supersimple.User, error) {
	ctx, collection := GoMongo("simple", "users")

	filter := bson.D{
		{"_id", id},
	}

	var u supersimple.User

	err := collection.FindOneAndDelete(ctx, filter).Decode(&u)
	if err != nil {
		log.Println("Error:", err)
	}
	return &u, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) OneUser(ctx context.Context, id *primitive.ObjectID, name *string) (*supersimple.User, error) {
	ctx, collection := GoMongo("simple", "users")

	var filter bson.D

	if id != nil {
		filter = bson.D{
			{"_id", id},
		}
	} else if name != nil {
		filter = bson.D{
			{"name", name},
		}
	} else if id != nil && name != nil {
		filter = bson.D{
			{"_id", id},
			{"name", name},
		}
	}

	var u supersimple.User

	err := collection.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		log.Println("Error:", err)
	}
	return &u, nil
}

func (r *queryResolver) Users(ctx context.Context) ([]*supersimple.User, error) {
	ctx, collection := GoMongo("simple", "users")

	var results []*supersimple.User

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Println(err)
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var elem supersimple.User
		err := cur.Decode(&elem)
		if err != nil {
			log.Println(err)
		}
		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		log.Println(err)
	}

	return results, nil
}
