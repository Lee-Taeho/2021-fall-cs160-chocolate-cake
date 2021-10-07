package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/middleware"
	"server/utils"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	AUTO_LOGOUT_TIME    = time.Minute * 10
	TIME_BEFORE_EXPIRED = time.Second * 30
)

var jwtKey = []byte("secret_key")

type Claims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func (h *Handlers) ExampleJsonReponse(w http.ResponseWriter, r *http.Request) {
	var courses []*middleware.Course

	course := new(middleware.Course)
	course.Name = "CS160"
	course.Department = "computer science"
	courses = append(courses, course)

	course2 := new(middleware.Course)
	course2.Name = "PHIL134"
	course2.Department = "philosophy"
	courses = append(courses, course2)

	fmt.Fprint(w, utils.Jsonify(courses))
}

func (h *Handlers) SaveNewUser(w http.ResponseWriter, r *http.Request) {
	log.Println("INFO [handlers/requestHandlers.go] Request to Save New User")
	var student middleware.Student

	if err := json.NewDecoder(r.Body).Decode(&student); err != nil {
		log.Printf("ERROR [handlers/requestHandlers.go] Couldn't get data: %s\n", err.Error())
		http.Error(w, "Couldn't get data", http.StatusInternalServerError)
		return
	}

	if err := h.db.CreateNewStudent(student); err != nil {
		log.Printf("ERROR [handlers/requestHandlers.go] Couldn't Save New User: %s\n", err.Error())
		http.Error(w, "Couldn't Save New User", http.StatusInternalServerError)
		return
	}
	log.Println("INFO [handlers/requestHandlers.go] Successfully Saved New User")
	// fmt.Fprint(w, "Thanks For signing up for CheckingN!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) LoginRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("INFO [handlers/requestHandlers.go] Request to Log In")
	var login middleware.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&login); err != nil {
		log.Printf("ERROR [handlers/requestHandlers.go] Couldn't get data: %s\n", err.Error())
		http.Error(w, "Couldn't get data", http.StatusInternalServerError)
		return
	}

	if found := h.db.FindStudent(login); !found {
		log.Println("INFO [handlers/requestHandlers.go] Failed Log In")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Println("INFO [handlers/requestHandlers.go] Log In Successful")

	expirationTime := time.Now().Add(AUTO_LOGOUT_TIME)

	claims := &Claims{
		Email: login.Email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	h.LoginRequestSuccess(w, r, tokenString)
}

func (h *Handlers) LoginRequestSuccess(w http.ResponseWriter, r *http.Request, tokenString string) {
	r.AddCookie(
		&http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: time.Now().Add(AUTO_LOGOUT_TIME),
		})
	if h.tokenValid(w, r) {
		// w.Header().Set("Access-Control-Allow-Credentials", "true")
		fmt.Fprint(w, "<h1>Successful Login!<h1>")
	}
}

// func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
// 	cookie, err := r.Cookie("token")
// 	if err != nil {
// 		if err == http.ErrNoCookie {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	tokenStr := cookie.Value

// 	claims := &Claims{}

// 	tkn, err := jwt.ParseWithClaims(tokenStr, claims,
// 		func(t *jwt.Token) (interface{}, error) {
// 			return jwtKey, nil
// 		})

// 	if err != nil {
// 		if err == jwt.ErrSignatureInvalid {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			return
// 		}
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	if !tkn.Valid {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		return
// 	}

// 	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > TIME_BEFORE_EXPIRED {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	expirationTime := time.Now().Add(AUTO_LOGOUT_TIME)

// 	claims.ExpiresAt = expirationTime.Unix()

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tokenString, err := token.SignedString(jwtKey)

// 	if err != nil {
// 		w.WriteHeader(http.StatusInternalServerError)
// 		return
// 	}

// 	http.SetCookie(w,
// 		&http.Cookie{
// 			Name:    "token",
// 			Value:   tokenString,
// 			Expires: expirationTime,
// 		})

// }
