package supersimple

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateAuthor(ctx context.Context, input NewAuthor) (*Author, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateAuthor(ctx context.Context, id primitive.ObjectID, dateOfDeath time.Time) (*Author, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteAuthor(ctx context.Context, id primitive.ObjectID) (*Author, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateBook(ctx context.Context, input NewBook) (*Book, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateBook(ctx context.Context, id primitive.ObjectID, outOfPrint *bool) (*Book, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteBook(ctx context.Context, id primitive.ObjectID) (*Book, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) OneAuthor(ctx context.Context, id *primitive.ObjectID, first *string, last *string, dateOfBirth *time.Time, alive *bool) (*Author, error) {
	panic("not implemented")
}
func (r *queryResolver) OneBook(ctx context.Context, id *primitive.ObjectID, title *string, genre *string, description *string, publisher *string, outOfPrint *bool) (*Book, error) {
	panic("not implemented")
}
func (r *queryResolver) Authors(ctx context.Context) ([]*Author, error) {
	panic("not implemented")
}
func (r *queryResolver) Books(ctx context.Context) ([]*Book, error) {
	panic("not implemented")
}
