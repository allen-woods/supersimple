package supersimple

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/allen-woods/supersimple/auth"
	"github.com/allen-woods/supersimple/database"
	supersimple "github.com/allen-woods/supersimple/models"
	"go.mongodb.org/mongo-driver/bson"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func (r *mutationResolver) SignUp(ctx context.Context, input *supersimple.NewUser) (*supersimple.User, error) {
	// Create a client connected to a generic context.
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in signUp():", err)
	}

	// Defer an immediately invoked function expression that disconnects the client.
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from signUp():", err)
		}
	}()

	// Point to the collection "users" in database "simple".
	collection := *client.Database("simple").Collection("users")

	// Enforce unique email values in the collection "users".
	err = database.RequireUniqueEmailFields(&collection)
	if err != nil {
		log.Fatal("Failed to require unique email fields:", err)
	}

	// Hash and salt the incoming password.
	securePassword, err := auth.HashAndSalt(input.Password)
	if err != nil {
		return nil, err
	}

	// Prepare data for a new User.
	u := &supersimple.User{
		Email:    input.Email,
		Name:     input.Name,
		UserName: input.UserName,
		Password: securePassword,
	}

	// Insert the new User into our generic context.
	res, err := collection.InsertOne(context.TODO(), *u)
	if err != nil {
		return nil, err
	}

	// Store the inserted "id".
	id := res.InsertedID
	// Assert the type of "id" to "primitive.ObjectID"
	u.ID = id.(primitive.ObjectID)
	// Pass the hex string of "id" to be consumed by authentication middleware.
	auth.InsertUserID(id.(primitive.ObjectID).Hex())

	// Return the User and no error
	return u, nil
}

func (r *mutationResolver) LogInUser(ctx context.Context, input *supersimple.NewUser) (*supersimple.User, error) {
	panic("not implemented")
}
func (r *mutationResolver) LogOutUser(ctx context.Context, id primitive.ObjectID) (bool, error) {
	userID := auth.ForContext(ctx)
	fmt.Println("This would have logged out the user with the id:", userID)
	return true, nil
}
func (r *mutationResolver) DeleteAccount(ctx context.Context, id primitive.ObjectID) (bool, error) {
	panic("not implemented")
}

func (r *mutationResolver) CreateAuthor(ctx context.Context, input supersimple.NewAuthor) (*supersimple.Author, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in createAuthor():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from createAuthor():", err)
		}
	}()

	collection := *client.Database("simple").Collection("authors")

	a := &supersimple.Author{
		First:       input.First,
		Last:        input.Last,
		DateOfBirth: input.DateOfBirth,
		DateOfDeath: input.DateOfDeath,
	}

	res, err := collection.InsertOne(context.TODO(), *a)
	if err != nil {
		log.Println(err)
	}

	id := res.InsertedID
	a.ID = id.(primitive.ObjectID)

	return a, nil
}
func (r *mutationResolver) UpdateAuthor(ctx context.Context, id primitive.ObjectID, dateOfDeath time.Time) (*supersimple.Author, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in updateAuthor():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from updateAuthor():", err)
		}
	}()

	collection := *client.Database("simple").Collection("authors")

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

	err = collection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&a)
	if err != nil {
		log.Println("Error:", err)
	}

	return &a, nil
}
func (r *mutationResolver) DeleteAuthor(ctx context.Context, id primitive.ObjectID) (*supersimple.Author, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in deleteAuthor():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from deleteAuthor():", err)
		}
	}()

	collection := *client.Database("simple").Collection("authors")

	filter := bson.D{
		{"_id", id},
	}

	var a supersimple.Author

	err = collection.FindOneAndDelete(context.TODO(), filter).Decode(&a)
	if err != nil {
		log.Println("Error:", err)
	}

	return &a, nil
}
func (r *mutationResolver) CreateBook(ctx context.Context, input supersimple.NewBook) (*supersimple.Book, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in createBook():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from createBook():", err)
		}
	}()

	collection := *client.Database("simple").Collection("books")

	b := &supersimple.Book{
		AuthorIDs:   input.AuthorIDs,
		Title:       input.Title,
		Genre:       input.Genre,
		Description: input.Description,
		Publisher:   input.Publisher,
		OutOfPrint:  input.OutOfPrint,
	}

	res, err := collection.InsertOne(context.TODO(), *b)
	if err != nil {
		log.Println(err)
	}

	id := res.InsertedID
	b.ID = id.(primitive.ObjectID)

	return b, nil
}
func (r *mutationResolver) UpdateBook(ctx context.Context, id primitive.ObjectID, outOfPrint bool) (*supersimple.Book, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in updateBook():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from updateBook():", err)
		}
	}()

	collection := *client.Database("simple").Collection("books")

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

	err = collection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&b)
	if err != nil {
		log.Println("Error:", err)
	}
	return &b, nil
}
func (r *mutationResolver) DeleteBook(ctx context.Context, id primitive.ObjectID) (*supersimple.Book, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in deleteBook():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from deleteBook():", err)
		}
	}()

	collection := *client.Database("simple").Collection("books")

	filter := bson.D{
		{"_id", id},
	}

	var b supersimple.Book

	err = collection.FindOneAndDelete(context.TODO(), filter).Decode(&b)
	if err != nil {
		log.Println("Error:", err)
	}
	return &b, nil
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Me(ctx context.Context) (*supersimple.User, error) {
	panic("not implemented")
}
func (r *queryResolver) Users(ctx context.Context) ([]*supersimple.User, error) {
	panic("not implemented")
}

func (r *queryResolver) OneAuthor(ctx context.Context, id *primitive.ObjectID, first *string, last *string, dateOfBirth *time.Time, dateOfDeath *time.Time) (*supersimple.Author, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in oneAuthor():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from oneAuthor():", err)
		}
	}()

	collection := *client.Database("simple").Collection("authors")

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
		bson.D{
			{
				"$project", bson.D{
					{"books.authors", 0},
				},
			},
		},
	}

	opts := options.Aggregate()
	opts.SetMaxAwaitTime(time.Second * 3)
	opts.SetMaxTime(time.Second * 3)

	cur, err := collection.Aggregate(context.TODO(), pipeline, opts)
	if err != nil {
		log.Println("Error:", err)
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		if err := cur.Decode(&a); err != nil {
			log.Fatal(err)
		}
	}

	return &a, nil
}
func (r *queryResolver) OneBook(ctx context.Context, id *primitive.ObjectID, title *string, genre *string, description *string, publisher *string, outOfPrint *bool) (*supersimple.Book, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in oneBook():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from oneBook():", err)
		}
	}()

	collection := *client.Database("simple").Collection("books")

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
		bson.D{
			{
				"$project", bson.D{
					{"authors.books", 0},
				},
			},
		},
	}

	opts := options.Aggregate()
	opts.SetMaxAwaitTime(time.Second * 3)
	opts.SetMaxTime(time.Second * 3)

	cur, err := collection.Aggregate(context.TODO(), pipeline, opts)
	if err != nil {
		log.Println("Error:", err)
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		if err := cur.Decode(&b); err != nil {
			log.Fatal(err)
		}
	}

	return &b, nil
}
func (r *queryResolver) Authors(ctx context.Context) ([]*supersimple.Author, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in authors():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from authors():", err)
		}
	}()

	collection := *client.Database("simple").Collection("authors")

	var results []*supersimple.Author

	cur, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Println(err)
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
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
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in books():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from books():", err)
		}
	}()

	collection := *client.Database("simple").Collection("books")

	var results []*supersimple.Book

	cur, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Println(err)
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
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

// 5dd7cc447cde2a00fd8ecf4e straub
// 5dd7ccc37cde2a00fd8ecf4f king
