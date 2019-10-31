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
	ctx, collection := GoMongo("simple", "authors")

	a := &supersimple.Author{
		First:       input.First,
		Last:        input.Last,
		DateOfBirth: input.DateOfBirth,
		DateOfDeath: input.DateOfDeath,
	}

	res, err := collection.InsertOne(ctx, *a)
	if err != nil {
		log.Println(err)
	}
	id := res.InsertedID
	a.ID = id.(primitive.ObjectID)
	return a, nil
}
func (r *mutationResolver) UpdateAuthor(ctx context.Context, id primitive.ObjectID, dateOfDeath time.Time) (*supersimple.Author, error) {
	ctx, collection := GoMongo("simple", "authors")

	filter := bson.D{
		{"_id", id},
	}

	update := bson.D{
		{
			"$set", bson.D{
				{"dateOfDeath", dateOfDeath},
			},
		},
	}

	opts := options.FindOneAndUpdate()
	opts.SetReturnDocument(options.After)

	var a supersimple.Author

	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&a)
	if err != nil {
		log.Println("Error:", err)
	}
	return &a, nil
}
func (r *mutationResolver) DeleteAuthor(ctx context.Context, id primitive.ObjectID) (*supersimple.Author, error) {
	ctx, collection := GoMongo("simple", "authors")

	filter := bson.D{
		{"_id", id},
	}

	var a supersimple.Author

	err := collection.FindOneAndDelete(ctx, filter).Decode(&a)
	if err != nil {
		log.Println("Error:", err)
	}
	return &a, nil
}
func (r *mutationResolver) CreateBook(ctx context.Context, input supersimple.NewBook) (*supersimple.Book, error) {
	ctx, collection := GoMongo("simple", "books")

	b := &supersimple.Book{
		AuthorIDs:   input.AuthorIDs,
		Title:       input.Title,
		Genre:       input.Genre,
		Description: input.Description,
		Publisher:   input.Publisher,
		OutOfPrint:  input.OutOfPrint,
	}

	res, err := collection.InsertOne(ctx, *b)
	if err != nil {
		log.Println(err)
	}
	id := res.InsertedID
	b.ID = id.(primitive.ObjectID)
	return b, nil
}
func (r *mutationResolver) UpdateBook(ctx context.Context, id primitive.ObjectID, outOfPrint bool) (*supersimple.Book, error) {
	ctx, collection := GoMongo("simple", "books")

	filter := bson.D{
		{"_id", id},
	}

	update := bson.D{
		{
			"$set", bson.D{
				{"outOfPrint", outOfPrint},
			},
		},
	}

	opts := options.FindOneAndUpdate()
	opts.SetReturnDocument(options.After)

	var b supersimple.Book

	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&b)
	if err != nil {
		log.Println("Error:", err)
	}
	return &b, nil
}
func (r *mutationResolver) DeleteBook(ctx context.Context, id primitive.ObjectID) (*supersimple.Book, error) {
	ctx, collection := GoMongo("simple", "books")

	filter := bson.D{
		{"_id", id},
	}

	var b supersimple.Book

	err := collection.FindOneAndDelete(ctx, filter).Decode(&b)
	if err != nil {
		log.Println("Error:", err)
	}
	return &b, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) OneAuthor(ctx context.Context, id *primitive.ObjectID, first *string, last *string, dateOfBirth *time.Time, dateOfDeath *time.Time) (*supersimple.Author, error) {
	ctx, collection := GoMongo("simple", "authors")

	var filter bson.D

	if id != nil {
		filter = append(filter, bson.E{"_id", id})
	}
	if first != nil {
		filter = append(filter, bson.E{"first", first})
	}
	if last != nil {
		filter = append(filter, bson.E{"last", last})
	}
	if dateOfBirth != nil {
		filter = append(filter, bson.E{"dateOfBirth", dateOfBirth})
	}
	if dateOfDeath != nil {
		filter = append(filter, bson.E{"dateOfDeath", dateOfDeath})
	}

	var a supersimple.Author
	pipeline := []bson.D{
		bson.D{
			{
				"$match", filter,
			},
		},
		bson.D{
			{
				"$lookup", bson.D{
					{"from", "books"},
					{"localField", "_id"},
					{"foreignField", "author_id"},
					{"as", "books"},
				},
			},
		},
	}

	opts := options.Aggregate()
	opts.SetMaxAwaitTime(time.Second * 3)
	opts.SetMaxTime(time.Second * 3)

	cur, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		log.Println("Error:", err)
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		if err := cur.Decode(&a); err != nil {
			log.Fatal(err)
		}
		log.Printf("cursor is:\n\n%v", cur)
	}

	return &a, nil
}

func (r *queryResolver) OneBook(ctx context.Context, id *primitive.ObjectID, title *string, genre *string, description *string, publisher *string, outOfPrint *bool) (*supersimple.Book, error) {
	ctx, collection := GoMongo("simple", "books")

	var filter bson.D

	if id != nil {
		filter = append(filter, bson.E{"_id", id})
	}
	if title != nil {
		filter = append(filter, bson.E{"title", title})
	}
	if genre != nil {
		filter = append(filter, bson.E{"genre", genre})
	}
	if description != nil {
		filter = append(filter, bson.E{"description", description})
	}
	if publisher != nil {
		filter = append(filter, bson.E{"publisher", publisher})
	}
	if outOfPrint != nil {
		filter = append(filter, bson.E{"outOfPrint", outOfPrint})
	}

	var b supersimple.Book

	pipeline := []bson.D{
		bson.D{
			{
				"$match", filter,
			},
		},
		bson.D{
			{
				"$lookup", bson.D{
					{"from", "authors"},
					{"localField", "author_id"},
					{"foreignField", "_id"},
					{"as", "authors"},
				},
			},
		},
	}

	opts := options.Aggregate()
	opts.SetMaxAwaitTime(time.Second * 3)
	opts.SetMaxTime(time.Second * 3)

	cur, err := collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		log.Println("Error:", err)
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		if err := cur.Decode(&b); err != nil {
			log.Fatal(err)
		}
	}

	return &b, nil
}
func (r *queryResolver) Authors(ctx context.Context) ([]*supersimple.Author, error) {
	ctx, collection := GoMongo("simple", "authors")

	var results []*supersimple.Author

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Println(err)
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var elem supersimple.Author
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
func (r *queryResolver) Books(ctx context.Context) ([]*supersimple.Book, error) {
	ctx, collection := GoMongo("simple", "book")

	var results []*supersimple.Book

	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Println(err)
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var elem supersimple.Book
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
