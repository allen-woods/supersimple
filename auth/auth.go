package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/securecookie"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

var authenticatedUserID struct {
	id string
}

var deleteSessionAndCookie struct {
	flag bool
}

var validateCookie struct {
	hash []byte
	key  []byte
}

var sc *securecookie.SecureCookie

var userIDCtxKey = &contextKey{"userID"}

type contextKey struct {
	name string
}

/*	Source 1:
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

/* End Source 1 */

func RollHashesAndKeyes() error {
	var hf *os.File
	hfName := "./hash"

	_, err := os.Stat(hfName)
	if os.IsNotExist(err) {
		hf, err = os.Create(hfName)
		if err != nil {
			return err
		}

		err = hf.Chmod(0600)
		if err != nil {
			return err
		}

		err = hf.Chown(os.Getuid(), os.Getgid())
		if err != nil {
			return err
		}

		err = os.Chtimes(hfName, time.Now(), time.Now())
		if err != nil {
			return err
		}

	} else {
		hf, err = os.Open(hfName)
		if err != nil {
			return err
		}
	}

	// From here on, hf is a valid pointer to a File.

	defer hf.Close()

	fi, err := hf.Stat()
	if err != nil {
		return err
	}

	// Extract the modification time
	modTime := fi.ModTime()

	// Check to see if 24 hours have passed
	if time.Now().Sub(modTime) >= 24*time.Hour {

		// If 24+ hours, we must subtract the first 32 bytes and add 32 at the end.
		// We then conclude by updating Chtimes

	} else {

		// If less than 24 hours, we update once per hour

	}

	/*
		1. Create the file if it doesn't exist at all.	(done)
		2. Set chmod to 0600 (if not already set).	(done)
		3. Set chown to os.Getuid and os.Getgid (if not already set).	(done)

		In a seperate goroutine...

		4. Check to see if an hour has passed since last modification time.
				* If it has been at least an hour, append new cryptographic entry to the file.
				* If it has not been at least an hour, do not append.
		5. Check to see if the length of the bytes in the file are 24 * 32.
				* If that length exists, remove bytes 0-31 and append 32 new ones hourly.

		Meanwhile, in this func...

		6. Retrieve the latest 32 bytes of the data to use in encryption and decryption
	*/
	return nil
}

/*	Source 2:
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

func CheckPassword(hashedPassword []byte, rawPassword []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hashedPassword, rawPassword)
	if err != nil {
		return false, err
	}
	return true, nil
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

/* End Source 2 */

func ReadFromRedis(sessionID map[string]string) (string, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer client.Close()

	// Request only the value contained in the "userID"
	// field within the hash whose address is "sessionID",
	// formatted as a string.
	userID, err := client.Do("HGET", sessionID["sessionID"], "userID").String()
	if err != nil {
		return "", err
	}

	return userID, nil
}

func WriteToRedis(sessionID map[string]string, userID string, ttl time.Time) error {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	defer client.Close()

	// Optimistically write a hash containing the
	// sessionID and userID.
	//
	// (Automatically overwrites existing keys)
	_, err := client.Do("HMSET", sessionID["sessionID"], "userID", userID).Result()
	if err != nil {
		return err
	}

	_, err = client.ExpireAt(sessionID["sessionID"], ttl).Result()
	if err != nil {
		return err
	}

	return nil
}

func Middleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			SetEnvironmentVariables()

			var maxAge int
			var expiration time.Time

			// Check to see if we should erase everything.
			// NOTE: We will need to check again so we can remove
			// the authenticated user.
			del := deleteSessionAndCookie.flag
			if del == true {
				maxAge = 0
				expiration = time.Now().AddDate(0, 0, -1)
			} else {
				maxAge = 24 * 60 * 60
				expiration = time.Now().Add(24 * time.Hour)
			}

			// Check to see if a user was authenticated
			// in our signUp or logIn mutations.
			if len(authenticatedUserID.id) == 0 {
				// Allow unauthenticated visitors to access the resolvers.
				next.ServeHTTP(w, r)
				return
			} else if len(authenticatedUserID.id) > 0 {
				// Look for a cookie when an authenticated user has been found.
				c, err := r.Cookie("sid")

				// The cookie won't exist for a new sign up.
				if c == nil || err != nil {

					// Give our new visitor a unique identifier.
					sessionID := map[string]string{
						"sessionID": uuid.NewV4().String(),
					}

					// Persist their uuid to Redis for 24 hours, in seconds.
					err = WriteToRedis(sessionID, authenticatedUserID.id, expiration)
					if err != nil {
						log.Fatalln("Unable to write sessionID to Redis:", err)
					}

					// Extract the hash we just persisted.
					persistedID, err := ReadFromRedis(sessionID)
					if err != nil {
						log.Fatalln("Unable to read sessionID from Redis:", err)
					}

					// Confirm the data was persisted and not corrupted or dropped.
					if persistedID == authenticatedUserID.id {

						// Encrypt the uuid using AES-256 algorithm.
						encoded, err := sc.Encode("sid", sessionID)
						if err != nil {
							log.Fatalln("Failed to encode sessionID")
						}

						// Create the cookie to be set on the response.
						//
						// NOTE: In production, this cookie should have a Domain field specified.
						//
						cookie := &http.Cookie{
							Name:     "sid",
							Value:    encoded,
							HttpOnly: true,
							Path:     "/",
							MaxAge:   maxAge,
							Expires:  expiration,
						}

						// Set the cookie on response to the end user.
						http.SetCookie(w, cookie)

						// Update context and serve next handler.
						ctx := context.WithValue(r.Context(), userIDCtxKey, authenticatedUserID.id)
						r = r.WithContext(ctx)
						next.ServeHTTP(w, r)
						return
					}
				}
				// The cookie will exist if the user is logged in.
				//
				// This is because we are persisting sessions, meaning
				// we do not need to keep the Redis hash or session cookie
				// after the user logs out.
				cookie, err := r.Cookie("sid")

				if cookie == nil || err != nil {
					log.Fatalln("Unable to find cookie for logged in User:", err)
				}

				// Set aside a variable for receiving the sessionID from cookie.
				sessionID := make(map[string]string)

				err = sc.Decode("sid", cookie.Value, &sessionID)
				if err != nil {
					log.Fatalln("The session cookie has been tampered with:", err)
				}

				// Attempt to read the userID from Redis stored in the hash whose address is sessionID.
				userID, err := ReadFromRedis(sessionID)
				if err != nil {
					log.Fatalln("Unable to read from Redis:", err)
				}

				// Extend the lifespan of the session to 24 hours from now, in seconds.
				err = WriteToRedis(sessionID, userID, expiration)
				if err != nil {
					log.Fatalln("Unable to write sessionID to Redis:", err)
				}

				persistedID, err := ReadFromRedis(sessionID)
				if err != nil && del != true {
					log.Fatalln("Unable to read userID from Redis:", err)
				}

				// Update the lifespan of the session cookie to up to
				// 24 hours from now, in seconds.
				cookie.HttpOnly = true

				if del == true {
					cookie.Value = ""
				}

				cookie.Path = "/"
				cookie.MaxAge = maxAge
				cookie.Expires = expiration

				// Set the cookie on response to the end user.
				http.SetCookie(w, cookie)

				if del == true {
					// Conditionally remove the User's id if we are
					// deleting during logOutUser mutation.
					authenticatedUserID.id = ""

					// We have to reset the flag or subsequent logins
					// will fail.
					deleteSessionAndCookie.flag = false

				} else {

					// Otherwise, place the extracted userID into a value
					// of type that context can access.
					authenticatedUserID.id = persistedID

				}

				// Update context and serve next handler.
				ctx := context.WithValue(r.Context(), userIDCtxKey, authenticatedUserID.id)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
			}
		})
	}
}

func InsertUserID(userID string) {
	authenticatedUserID.id = userID
}

func SetLogOutFlag(value bool) {
	deleteSessionAndCookie.flag = value
}

func ForContext(ctx context.Context) string {
	raw, _ := ctx.Value(userIDCtxKey).(string)
	return raw
}
