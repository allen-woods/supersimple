package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis"
	"github.com/gorilla/securecookie"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

var uniqueUserID struct {
	id string
}

var validateCookie struct {
	hash []byte
	key  []byte
}

var sc *securecookie.SecureCookie

var uniqueUserIDCtxKey = &contextKey{"uuid"}

type contextKey struct {
	name string
}

/*	Source 2:
	https://stackoverflow.com/questions/32349807/how-can-i-generate-a-random-int-using-the-crypto-rand-package
*/
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
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

/* End Source 2 */

/*	Source 3:
	https://www.gorillatoolkit.org/pkg/securecookie
*/
func SetEnvironmentVariables() error {
	h := os.Getenv("COOKIE_HASH")
	if h == "" {
		hs, err := GenerateRandomString(24)
		if err != nil {
			return err
		}
		// If the environment variable does not exist,
		// we must initialize it with the hash string.
		os.Setenv("COOKIE_HASH", hs)
	}

	k := os.Getenv("COOKIE_KEY")
	if k == "" {
		ks, err := GenerateRandomString(24)
		if err != nil {
			return err
		}
		// If the environment variable does not exist,
		// we must initialize it with the key string.
		os.Setenv("COOKIE_KEY", ks)
	}

	validateCookie.hash = []byte(os.Getenv("COOKIE_HASH"))
	validateCookie.key = []byte(os.Getenv("COOKIE_KEY"))
	sc = securecookie.New(validateCookie.hash, validateCookie.key)

	return nil
}

func HashAndSalt(pw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func ReadSessionIDFromCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie("sid")
	if err != nil {
		return "", err
	}

	value := make(map[string]string)

	err = sc.Decode("sid", cookie.Value, &value)
	if err != nil {
		return "", err
	}

	validSessionID, err := uuid.FromString(value["sessionID"])
	if err != nil {
		return "", err
	}

	return validSessionID.String(), nil
}

/* End Source 3 */

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

func Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			SetEnvironmentVariables()

			if len(uniqueUserID.id) > 0 {
				encoded, err := sc.Encode("sessionID", uniqueUserID.id)
				if err != nil {
					log.Fatal("Failed to encode sessionID")
				}

				cookie := &http.Cookie{
					Name:     "sid",
					Value:    encoded,
					HttpOnly: true,
					Path:     "/",
					//Domain:   "127.0.0.1",
					MaxAge: 24 * 60 * 60,
				}

				http.SetCookie(w, cookie)
				fmt.Println("Cookie should have been set")
				uniqueUserID.id = ""
			}

			ctx := context.WithValue(r.Context(), uniqueUserIDCtxKey, uniqueUserID.id)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func TransferUUID(uuid string) {
	uniqueUserID.id = uuid
}
