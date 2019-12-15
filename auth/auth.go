package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
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

func Roll() error {
	err := RollFile(".hash")
	if err != nil {
		return err
	}

	err = RollFile(".key")
	if err != nil {
		return err
	}

	return nil
}

func RollFile(fName string) error {
	var f *os.File

	_, err := os.Stat(fName)
	if os.IsNotExist(err) {
		f, err = os.Create(fName)
		if err != nil {
			return err
		}

		err = f.Chown(os.Getuid(), os.Getgid())
		if err != nil {
			return err
		}

		err = f.Chmod(0600)
		if err != nil {
			return err
		}

		err = os.Chtimes(fName, time.Now(), time.Now())
		if err != nil {
			return err
		}
	} else {
		f, err = os.OpenFile(fName, os.O_RDWR, 0600)
		if err != nil {
			return err
		}
	}

	defer f.Close()

	fInfo, err := f.Stat()
	if err != nil {
		return err
	}

	// Capture the number of bytes in the file.
	n := fInfo.Size()

	// Create a variable of type []byte for storing the file's data later.
	var fData []byte

	if n >= 24*32 {
		// We have 24 hours worth of data stored in the file.
		// Initialize the length of fData to n - 32.
		// We do this because we are trimming off the first 32 bytes.
		fData = make([]byte, n-32)

		fmt.Println("/ // / // / 24 Hours / // / // /")

		lenBytes, err := f.ReadAt(fData, 32)
		if lenBytes != int(n-32) || err != nil {
			return errors.New("Corruption of data in rolling encryption file:\nunexpected number of bytes.")
		}
	} else {
		// Initialize the length of fData to n.
		// We do this because we are only appending bytes, not removing bytes.
		fData = make([]byte, n)

		lenBytes, err := f.Read(fData)
		if lenBytes != int(n) || err != nil {
			return err //errors.New("Corruption of data in rolling encryption file:\nunexpected number of bytes.")
		}
	}

	// Create our random string.
	s, err := GenerateRandomString(24)
	if err != nil {
		return err
	}

	//fmt.Printf("The data contains:\n%v\n\nand is %d bytes long.\n\n", fData, len(fData))

	//fmt.Println("APPENDING...")

	// Append the random string to fData.
	fData = append(fData, []byte(s)...)

	//fmt.Printf("\nThe data NOW contains:\n%v\n\nand is %d bytes long.\n\n", fData, len(fData))

	// Overwrite the contents of the file.
	_, err = f.WriteAt(fData, 0)
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}
	// if numBytes != len(fData) {
	// 	return fmt.Errorf("Number of bytes written: %d\nNumber of bytes expected: %d\n\n", numBytes, len(fData))
	// } else if err != nil {
	// 	return err
	// }

	// Send the newest entry to the unexported variables.
	// The newest entry is the last 32 bytes and must be sent to the correct var.
	switch fName {
	case ".hash":
		validateCookie.hash = fData[len(fData)-32:]
		//fmt.Printf("Appended byte slice:\n%s\nTarget file: %s\n\n", string(validateCookie.hash), fName)
	case ".key":
		validateCookie.key = fData[len(fData)-32:]
		//fmt.Printf("Appended byte slice:\n%s\nTarget file: %s\n\n", string(validateCookie.key), fName)
	}

	// If we have made it through, everything worked correctly.
	return nil
}

/*	Source 2:
	https://www.gorillatoolkit.org/pkg/securecookie
*/

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
		// This is where we need to check all possible hashes and keys
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

						sc = securecookie.New(validateCookie.hash, validateCookie.key)

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
