package supersimple

import (
	"context"
	"log"
	"time"

	supersimple "github.com/allen-woods/supersimple/models"
	// "go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

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

func (r *mutationResolver) CreateAuthor(ctx context.Context, input supersimple.NewAuthor) (*supersimple.Author, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateAuthor(ctx context.Context, id primitive.ObjectID, dateOfDeath time.Time) (*supersimple.Author, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteAuthor(ctx context.Context, id primitive.ObjectID) (*supersimple.Author, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateBook(ctx context.Context, input supersimple.NewBook) (*supersimple.Book, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateBook(ctx context.Context, id primitive.ObjectID, outOfPrint *bool) (*supersimple.Book, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteBook(ctx context.Context, id primitive.ObjectID) (*supersimple.Book, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) OneAuthor(ctx context.Context, id *primitive.ObjectID, first *string, last *string, dateOfBirth *time.Time, alive *bool) (*supersimple.Author, error) {
	panic("not implemented")
}
func (r *queryResolver) OneBook(ctx context.Context, id *primitive.ObjectID, title *string, genre *string, description *string, publisher *string, outOfPrint *bool) (*supersimple.Book, error) {
	panic("not implemented")
}
func (r *queryResolver) Authors(ctx context.Context) ([]*supersimple.Author, error) {
	panic("not implemented")
}
func (r *queryResolver) Books(ctx context.Context) ([]*supersimple.Book, error) {
	panic("not implemented")
}
