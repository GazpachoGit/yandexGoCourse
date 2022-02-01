package middlewares

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"net/http"

	uuid "github.com/nu7hatch/gouuid"
)

type contextKey string

func (c contextKey) String() string {
	return "mypackage context key " + string(c)
}

var contextKeyUser = contextKey("user")

var secretkey = []byte("$ecret key")

type userCookie struct {
	token *http.Cookie
	User  string
	New   bool
}
type UserInfo struct {
	UserID string
	New    bool
}

func CockieHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var currentUC *userCookie
		cookie, err := r.Cookie("token")
		if err != nil {
			currentUC, err = setCookie()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			currentUC, err = getUser(cookie)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		log.Println("is user new: ", currentUC.New)
		if currentUC.New {
			http.SetCookie(w, currentUC.token)
		}

		ctx := context.WithValue(r.Context(), contextKeyUser, currentUC.User)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func setCookie() (*userCookie, error) {
	userGUID, err := uuid.NewV4()
	user := userGUID.String()
	if err != nil {
		return nil, err
	}
	h := hmac.New(sha256.New, secretkey)
	h.Write([]byte(user))
	token := h.Sum(nil)

	tockenEncode := hex.EncodeToString(token)
	userEncode := hex.EncodeToString([]byte(user))

	cookie := &http.Cookie{
		Path:   "/",
		Name:   "token",
		Value:  userEncode + tockenEncode,
		MaxAge: 300,
	}
	return &userCookie{cookie, user, true}, nil
}

func getUser(cookie *http.Cookie) (*userCookie, error) {
	value := cookie.Value
	if value != "" {
		decodedValue, err := hex.DecodeString(value)
		if err == nil {
			user := decodedValue[:36]
			h := hmac.New(sha256.New, secretkey)
			h.Write(user)
			sign := h.Sum(nil)
			if hmac.Equal(sign, decodedValue[36:]) {
				return &userCookie{token: cookie, User: string(user), New: false}, nil
			}
		}
	}
	return setCookie()
}

func GetUserFromCxt(ctx context.Context) (string, bool) {
	user, ok := ctx.Value(contextKeyUser).(string)
	return user, ok
}
