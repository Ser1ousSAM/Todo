package routing

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"main/db"
	"net/http"
)

type DBInterface interface {
	CreateUser(user db.User) (db.User, error)
	Auth(user db.User) (db.User, error)
}

type Route struct {
	DB DBInterface
}

type CustomError struct {
	Message string `json:"message"`
}

func Error(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")

	ce := CustomError{
		Message: msg,
	}

	res, errM := json.Marshal(ce)
	if errM != nil {
		zap.S().Errorw("marshal", "error", errM)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Write(res)
}

//func loggingMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		log.Println(r.RequestURI)
//		next.ServeHTTP(w, r)
//	})
//}

func NewRouter(ro Route) *mux.Router {
	router := mux.NewRouter()
	//router.Use(loggingMiddleware)

	router.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var user db.User
		if errCU := UnmarshalBody(r.Body, &user); errCU != nil {
			http.Error(w, errCU.Error(), http.StatusInternalServerError)
			return
		}

		user, errCreate := ro.DB.CreateUser(user)
		if errCreate != nil {
			http.Error(w, errCreate.Error(), http.StatusInternalServerError)
			return
		}

		token, errToken := GenToken(user)
		if errToken != nil {
			http.Error(w, errToken.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(token)

	}).Methods("POST")

	router.HandleFunc("/auth", func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var user db.User
		if errAU := UnmarshalBody(r.Body, &user); errAU != nil {
			http.Error(rw, errAU.Error(), http.StatusInternalServerError)
			return
		}

		user, errAuth := ro.DB.Auth(user)
		if errAuth != nil {
			Error(rw, errAuth.Error(), http.StatusUnauthorized)
			return
		}

		token, err := GenToken(user)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(token)

	}).Methods("POST")

	return router
}
func GenToken(user db.User) ([]byte, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"login": user.Login,
	})
	t, err := token.SigningString()
	if err != nil {
		return nil, fmt.Errorf("gen token: %w", err)
	}

	type TokenResp struct {
		Token string `json:"token"`
	}

	tok := TokenResp{
		Token: t,
	}

	data, err := json.Marshal(tok)
	if err != nil {
		return nil, fmt.Errorf("marshal token: %w", err)
	}

	return data, nil
}
func UnmarshalBody(r io.Reader, v interface{}) error {
	//check why it's deprecated
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	if errJson := json.Unmarshal(data, &v); errJson != nil {
		return fmt.Errorf("unmarshal: %w", errJson)
	}

	return nil
}
