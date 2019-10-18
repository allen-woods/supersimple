package supersimple

import (
	"errors"
	"io"
	"log"

	"github.com/99designs/gqlgen/graphql"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

type NewUser struct {
	Name string
}

type User struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
}

func MarshalID(id primitive.ObjectID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		json, err := id.MarshalJSON()

		if err != nil {
			log.Fatal("Error:", err)
		}

		io.WriteString(w, string(json))
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
		err = id.UnmarshalJSON([]byte(json))
	}

	return id, err
}
