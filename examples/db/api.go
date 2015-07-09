package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type server struct {
	db *sql.DB
}

type user struct {
	ID       int64  `json:"-"`
	Username string `json:"username"`
	Email    string `json:"-"`
}

func (s *server) users(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		fail(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var users []*user
	rows, err := s.db.Query("SELECT id, email, username FROM users")
	defer rows.Close()
	switch err {
	case nil:
		for rows.Next() {
			user := &user{}
			if err := rows.Scan(&user.ID, &user.Email, &user.Username); err != nil {
				fail(w, fmt.Sprintf("failed to scan an user: %s", err), http.StatusInternalServerError)
				return
			}
			users = append(users, user)
		}
		if len(users) == 0 {
			users = make([]*user, 0) // an empty array in this case
		}
	default:
		fail(w, fmt.Sprintf("failed to fetch users: %s", err), http.StatusInternalServerError)
		return
	}

	data := struct {
		Users []*user `json:"users"`
	}{Users: users}

	ok(w, data)
}

func main() {
	db, err := sql.Open("mysql", "root@/godog")
	if err != nil {
		panic(err)
	}
	s := &server{db: db}
	http.HandleFunc("/users", s.users)
	http.ListenAndServe(":8080", nil)
}

// fail writes a json response with error msg and status header
func fail(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")

	data := struct {
		Error string `json:"error"`
	}{Error: msg}

	resp, _ := json.Marshal(data)
	w.WriteHeader(status)

	fmt.Fprintf(w, string(resp))
}

// ok writes data to response with 200 status
func ok(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if s, ok := data.(string); ok {
		fmt.Fprintf(w, s)
		return
	}

	resp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fail(w, "oops something evil has happened", 500)
		return
	}

	fmt.Fprintf(w, string(resp))
}
