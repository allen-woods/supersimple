package supersimple

import (
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
	hex, ok := v.(string)
	if !ok {
		log.Fatal("ids must be strings")
	}

	objID, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		log.Fatal("Error:", err)
	}

	return objID, err
}
