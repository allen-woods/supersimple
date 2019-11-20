package supersimple

import (
	"errors"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

type NewUser struct {
	Email    string
	Name     string
	UserName string
	Password string
}

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Email    string
	Name     string
	UserName string
	Password string
}

type NewAuthor struct {
	First       string
	Last        string
	DateOfBirth time.Time
	DateOfDeath time.Time
}

type Author struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	First       string
	Last        string
	DateOfBirth time.Time
	DateOfDeath time.Time
	Books       []Book
}

type NewBook struct {
	AuthorIDs   []primitive.ObjectID `bson:"author_id,omitempty"`
	Title       string
	Genre       string
	Description string
	Publisher   string
	OutOfPrint  bool
}

type Book struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty"`
	AuthorIDs   []primitive.ObjectID `bson:"author_id,omitempty"`
	Authors     []Author
	Title       string
	Genre       string
	Description string
	Publisher   string
	OutOfPrint  bool
}

func MarshalID(id primitive.ObjectID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		json, err := id.MarshalJSON()

		if err != nil {
			log.Fatal("Error:", err)
		}

		io.WriteString(w, strconv.Quote(string(json)))
	})
}

func UnmarshalID(v interface{}) (primitive.ObjectID, error) {
	var id primitive.ObjectID
	var err error
	err = nil

	json, ok := v.(string)

	if !ok {
		err = errors.New("ids must be strings")
	} else {
		err = id.UnmarshalJSON([]byte(strconv.Quote(json)))
	}

	return id, err
}
