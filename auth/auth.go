package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"

	db "github.com/allen-woods/supersimple/database"
	"github.com/go-redis/redis"
	"github.com/gorilla/securecookie"
	"github.com/rbcervilla/redisstore"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
)

var cookieHash []byte
var cookieKey []byte
var scInstance *securecookie.SecureCookie

var store *redisstore.RedisStore
var userCtxKey = &contextKey{name: "user"}

type contextKey struct {
	name string
}

type User struct {
	Email    string
	Name     string
	UserName string
	IsAdmin  bool
}

/*	Source 1:
	https://stackoverflow.com/questions/32349807/how-can-i-generate-a-random-int-using-the-crypto-rand-package
*/
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	// Return a slice of byte containing the
	// cryptographically random string.
	return b, nil
}

func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

/* End Source 1 */

/*	Source 2:
	https://www.gorillatoolkit.org/pkg/securecookie
*/
func SetEnvironmentVariables() error {
	h := os.Getenv("COOKIE_HASH")
	if h == "" {
		hs, err := GenerateRandomString(32)
		if err != nil {
			return err
		}
		// If the environment variable does not exist,
		// we must initialize it with the hash string.
		os.Setenv("COOKIE_HASH", hs)
	}

	k := os.Getenv("COOKIE_KEY")
	if k == "" {
		ks, err := GenerateRandomString(32)
		if err != nil {
			return err
		}
		// If the environment variable does not exist,
		// we must initialize it with the key string.
		os.Setenv("COOKIE_KEY", ks)
	}

	cookieHash = []byte(os.Getenv("COOKIE_HASH"))
	cookieKey = []byte(os.Getenv("COOKIE_KEY"))

	return nil
}

func SetCookieHandler(w http.ResponseWriter, r *http.Request, sID string) error {
	// This function must be passed a valid uuid.NewV4().String()
	// as its sID argument.
	scInstance = securecookie.New(cookieHash, cookieKey)

	value := map[string]string{
		"sessionID": sID,
	}

	encoded, err := scInstance.Encode("sid", value)
	if err != nil {
		return err
	}

	cookie := &http.Cookie{
		Name:  "sid",
		Value: encoded,
		Path:  "/",
	}

	http.SetCookie(w, cookie)

	return nil
}

func ReadCookieHandler(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		return "", err
	}

	value := make(map[string]string)

	err = scInstance.Decode("sid", cookie.Value, &value)
	if err != nil {
		return "", err
	}

	// This function validates the uuid as genuine before
	// returning it.
	validSessionID, err := uuid.FromString(value["sessionID"])
	if err != nil {
		return "", err
	}

	return validSessionID.String(), nil
}

/* End Source 2 */

func ReadFromRedis(sessionID string) (string, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer client.Close()

	// Request only the value contained in the "userID"
	// field within the hash whose address is "sessionID",
	// formatted as a string.
	userID, err := client.Do("HGET", sessionID, "userID").String()
	if err != nil {
		return "", err
	}

	return userID, nil
}

func WriteToRedis(sessionID string, userID string) error {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer client.Close()

	// Optimistically write a hash containing the
	// sessionID and userID.
	//
	// (Automatically overwrites existing keys)
	_, err := client.Do("HMSET", sessionID, "userID", userID).Result()
	if err != nil {
		return err
	}

	return nil
}

func validateAndGetUserID(w http.ResponseWriter, r *http.Request) (string, error) {
	/* In this function:

	uuid.NewV4().String()

	- Use sessions to validate the cookie.
	- Use cookie value to find session in Redis.
	- Use session value to retrieve user id.

	*/
	sessionID, err := ReadCookieHandler(w, r)
	if err != nil {
		return "", err
	}

	userID, err := ReadFromRedis(sessionID)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func getUserByID(userID string) *User {
	_, collection := db.GoMongo("simple", "users")

	user := collection.FindOne(context.Background(), bson.D{{"_id", userID}})

	return &user
}

func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("sid")

			if err != nil || c == nil {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := validateAndGetUserID(w, r)
			if err != nil {
				http.Error(w, "Invalid cookie", http.StatusForbidden)
				return
			}

			user := getUserByID(userID)

			ctx := context.WithValue(r.Context(), userCtxKey, user)

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func ForContext(ctx context.Context) *User {
	raw, _ := ctx.Value(userCtxKey).(*User)
	return raw
}
