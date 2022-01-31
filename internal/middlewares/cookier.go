package middlewares

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

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

		ctx := context.WithValue(r.Context(), "user", currentUC.User)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
func setCookie() (*userCookie, error) {
	user, err := RandBytes()
	if err != nil {
		return nil, err
	}
	h := hmac.New(sha256.New, secretkey)
	h.Write([]byte(user))
	token := h.Sum(nil)

	tockenEncode := hex.EncodeToString(token)
	userEncode := hex.EncodeToString([]byte(user))

	cookie := &http.Cookie{
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
			user := decodedValue[:19]
			h := hmac.New(sha256.New, secretkey)
			h.Write(user)
			sign := h.Sum(nil)
			if hmac.Equal(sign, decodedValue[19:]) {
				return &userCookie{token: cookie, User: string(user), New: false}, nil
			}
		}
	}
	return setCookie()
}

func RandBytes() (string, error) {
	b := rand.Int()
	return strconv.Itoa(b), nil
}
