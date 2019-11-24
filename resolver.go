package supersimple

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/allen-woods/supersimple/auth"
	"github.com/allen-woods/supersimple/database"
	supersimple "github.com/allen-woods/supersimple/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

// Resolver provides the root type used by our resolvers.
type Resolver struct{}

// Mutation provides the root for our mutation resolvers.
func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

// Query provides the root for our query resolvers.
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// mutationResolver provides the root type specific to our mutation resolvers.
type mutationResolver struct{ *Resolver }

// SignUp provides a mutation for creating a new user.
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

func (r *mutationResolver) LogInUser(ctx context.Context, email string, password string) (*supersimple.User, error) {
	// Look to see if the user is already logged in
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser != "" {
		err := errors.New("User is already logged in")
		return nil, err
	}

	// Connect the client.
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in logInUser():", err)
	}

	// Defer disconnection of the client.
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from logInUser():", err)
		}
	}()

	// Point to the "users" collection.
	collection := *client.Database("simple").Collection("users")

	// Search using the email, which is required to be unique.
	filter := bson.D{
		bson.E{Key: "email", Value: email},
	}

	// Set the options for the FindOne operation.
	opts := options.FindOne()

	// Create a User struct where the data will be decoded.
	var u supersimple.User

	// Search for the User.
	err = collection.FindOne(context.TODO(), filter, opts).Decode(&u)
	if err != nil {
		log.Println("Error:", err)
	}

	// Confirm the password matches.
	ok, err := auth.CheckPassword([]byte(u.Password), []byte(password))
	if !ok {
		err := errors.New("Incorrect email or password")
		return nil, err
	}

	// Insert the id of the User to be consumed by
	// authentication middleware.
	auth.InsertUserID(u.ID.Hex())

	// Return a pointer to the User and no error.
	return &u, nil
}
func (r *mutationResolver) LogOutUser(ctx context.Context) (bool, error) {
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("User is already logged out")
		return false, err
	}

	// Tell the authentication middleware to delete the
	// session from Redis and to delete the "sid" cookie.
	auth.SetLogOutFlag(true)

	return true, nil
}
func (r *mutationResolver) DeleteAccount(ctx context.Context, id primitive.ObjectID, confirmDelete bool) (bool, error) {
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("User is already logged out")
		return false, err
	}

	authenticatedUser, err := primitive.ObjectIDFromHex(loggedInUser)

	if err != nil {
		err := errors.New("Unable to generate authenticatedUser in deleteAccount()")
		return false, err
	}

	// Check to make sure the information is fully confirmed.
	if authenticatedUser == id && confirmDelete == true {
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
		if err != nil {
			log.Fatal("Unable to create client in deleteAccount():", err)
		}

		defer func() {
			if err = client.Disconnect(context.TODO()); err != nil {
				log.Fatal("Unable to disconnect client from deleteAccount():", err)
			}
		}()

		collection := *client.Database("simple").Collection("users")

		filter := bson.D{
			bson.E{Key: "_id", Value: id},
		}

		var u supersimple.User

		err = collection.FindOneAndDelete(context.TODO(), filter).Decode(&u)
		if err != nil {
			err := errors.New("Unable to delete User in deleteAccount()")
			return false, err
		}

		auth.SetLogOutFlag(true)

		return true, nil
	}
	err = errors.New("Authentication or confirmation failure")
	return false, err
}

func (r *mutationResolver) CreateAuthor(ctx context.Context, input supersimple.NewAuthor) (*supersimple.Author, error) {
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("Access denied")
		return nil, err
	}

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
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("Access denied")
		return nil, err
	}

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
		bson.E{Key: "_id", Value: id},
	}

	update := bson.D{
		bson.E{
			Key: "$set", Value: bson.D{
				bson.E{Key: "dateOfDeath", Value: dateOfDeath},
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
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("Access denied")
		return nil, err
	}

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
		bson.E{Key: "_id", Value: id},
	}

	var a supersimple.Author

	err = collection.FindOneAndDelete(context.TODO(), filter).Decode(&a)
	if err != nil {
		log.Println("Error:", err)
	}

	return &a, nil
}
func (r *mutationResolver) CreateBook(ctx context.Context, input supersimple.NewBook) (*supersimple.Book, error) {
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("Access denied")
		return nil, err
	}

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
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("Access denied")
		return nil, err
	}

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
		bson.E{Key: "_id", Value: id},
	}

	update := bson.D{
		bson.E{
			Key: "$set", Value: bson.D{
				bson.E{Key: "outOfPrint", Value: outOfPrint},
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
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("Access denied")
		return nil, err
	}

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
		bson.E{Key: "_id", Value: id},
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
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("User is logged out")
		return nil, err
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in deleteBook():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from deleteBook():", err)
		}
	}()

	// Point to the "users" collection.
	collection := *client.Database("simple").Collection("users")

	// Create a User struct to decode bson into.
	var u supersimple.User

	// Format our loggedInUser as type primitive.ObjectID.
	id, err := primitive.ObjectIDFromHex(loggedInUser)
	if err != nil {
		err := errors.New("Unable to format loggedInUser in me()")
		return nil, err
	}

	// Create an aggregation pipeline so we can prevent
	// projection of the password field.
	pipeline := []bson.D{
		bson.D{
			bson.E{Key: "$match", Value: bson.D{
				bson.E{
					Key: "_id", Value: bson.D{
						bson.E{Key: "$in", Value: bson.A{id}},
					},
				}},
			},
		},
		bson.D{
			bson.E{
				Key: "$project", Value: bson.D{
					bson.E{Key: "password", Value: 0},
				},
			},
		},
	}

	// Declare options for the pipeline.
	opts := options.Aggregate()
	opts.SetMaxAwaitTime(time.Second * 3)
	opts.SetMaxTime(time.Second * 3)

	// Return the cursor of our matching User, if any.
	cur, err := collection.Aggregate(context.TODO(), pipeline, opts)
	if err != nil {
		log.Println("Error:", err)
	}

	// Defer the closure of the cursor.
	defer cur.Close(context.TODO())

	// Iterate through the cursor so long as there is a
	// next element.
	for cur.Next(context.TODO()) {
		if err := cur.Decode(&u); err != nil {
			log.Fatal(err)
		}
	}

	// Return the User and no error.
	return &u, nil
}
func (r *queryResolver) Users(ctx context.Context) ([]*supersimple.User, error) {
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("Access denied")
		return nil, err
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Unable to create client in deleteBook():", err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Unable to disconnect client from deleteBook():", err)
		}
	}()

	// Point to the "users" collection.
	collection := *client.Database("simple").Collection("users")

	// Create a User struct to decode bson into.
	var results []*supersimple.User

	// Format our loggedInUser as type primitive.ObjectID.
	id, err := primitive.ObjectIDFromHex(loggedInUser)
	if err != nil {
		err := errors.New("Unable to format loggedInUser in users()")
		return nil, err
	}

	// Create an aggregation pipeline so we can ignore the
	// authenticated User and prevent projection of the
	// password fields.
	pipeline := []bson.D{
		bson.D{
			bson.E{Key: "$match", Value: bson.D{
				bson.E{
					Key: "_id", Value: bson.D{
						bson.E{Key: "$nin", Value: bson.A{id}},
					},
				}},
			},
		},
		bson.D{
			bson.E{
				Key: "$project", Value: bson.D{
					bson.E{Key: "password", Value: 0},
				},
			},
		},
	}

	// Declare options for the pipeline.
	opts := options.Aggregate()
	opts.SetMaxAwaitTime(time.Second * 3)
	opts.SetMaxTime(time.Second * 3)

	// Return the cursor of our matching User, if any.
	cur, err := collection.Aggregate(context.TODO(), pipeline, opts)
	if err != nil {
		log.Println("Error:", err)
	}

	// Defer the closure of the cursor.
	defer cur.Close(context.TODO())

	// Iterate through the cursor so long as there is a
	// next element.
	for cur.Next(context.TODO()) {
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
func (r *queryResolver) OneAuthor(ctx context.Context, id *primitive.ObjectID, first *string, last *string, dateOfBirth *time.Time, dateOfDeath *time.Time) (*supersimple.Author, error) {
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("Access denied")
		return nil, err
	}

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
		filter = append(filter, bson.E{Key: "_id", Value: id})
	}
	if first != nil {
		filter = append(filter, bson.E{Key: "first", Value: first})
	}
	if last != nil {
		filter = append(filter, bson.E{Key: "last", Value: last})
	}
	if dateOfBirth != nil {
		filter = append(filter, bson.E{Key: "dateOfBirth", Value: dateOfBirth})
	}
	if dateOfDeath != nil {
		filter = append(filter, bson.E{Key: "dateOfDeath", Value: dateOfDeath})
	}

	var a supersimple.Author

	pipeline := []bson.D{
		bson.D{
			bson.E{
				Key: "$match", Value: filter,
			},
		},
		bson.D{
			bson.E{
				Key: "$lookup", Value: bson.D{
					bson.E{Key: "from", Value: "books"},
					bson.E{Key: "localField", Value: "_id"},
					bson.E{Key: "foreignField", Value: "author_id"},
					bson.E{Key: "as", Value: "books"},
				},
			},
		},
		bson.D{
			bson.E{
				Key: "$project", Value: bson.D{
					bson.E{Key: "books.authors", Value: 0},
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
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("Access denied")
		return nil, err
	}

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
		filter = append(filter, bson.E{Key: "_id", Value: id})
	}
	if title != nil {
		filter = append(filter, bson.E{Key: "title", Value: title})
	}
	if genre != nil {
		filter = append(filter, bson.E{Key: "genre", Value: genre})
	}
	if description != nil {
		filter = append(filter, bson.E{Key: "description", Value: description})
	}
	if publisher != nil {
		filter = append(filter, bson.E{Key: "publisher", Value: publisher})
	}
	if outOfPrint != nil {
		filter = append(filter, bson.E{Key: "outOfPrint", Value: outOfPrint})
	}

	var b supersimple.Book

	pipeline := []bson.D{
		bson.D{
			bson.E{
				Key: "$match", Value: filter,
			},
		},
		bson.D{
			bson.E{
				Key: "$lookup", Value: bson.D{
					bson.E{Key: "from", Value: "authors"},
					bson.E{Key: "localField", Value: "author_id"},
					bson.E{Key: "foreignField", Value: "_id"},
					bson.E{Key: "as", Value: "authors"},
				},
			},
		},
		bson.D{
			bson.E{
				Key: "$project", Value: bson.D{
					bson.E{Key: "authors.books", Value: 0},
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
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("Access denied")
		return nil, err
	}

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
	// Look to see if the user is already logged out.
	loggedInUser := auth.ForContext(ctx)
	if loggedInUser == "" {
		err := errors.New("Access denied")
		return nil, err
	}

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
