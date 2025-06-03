package main

import (
	"fmt"
	"net/http"
	"net/http/cgi"
	"os/exec"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("sVnOd8XxdyH2M2q0X0SphbaAmK81cKM2")

func main() {
	err := cgi.Serve(http.HandlerFunc(handler))
	if err != nil {
		fmt.Println("Failed to serve CGI request")
	}
}

func isAdmin(login string, w http.ResponseWriter) bool {
	req := "SELECT login FROM adminsP;"
	output, err := send_sql_request(req)
	if err != nil {
		fmt.Fprintln(w, "Ошибка при попытке получить логины админов")
	}

	admins := strings.Split(output, "\n")[2:]
	admins = admins[:len(admins)-1]

	//fmt.Fprintln(w, admins)
	for _, admin := range admins {
		if admin == login {
			return true
		}
	}
	return false
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Ivalid request method", http.StatusBadRequest)
		return
	}

	flag, login := check_JWT_token(r, w)
	if !flag || !isAdmin(login, w) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	user_id := r.FormValue("user_id")
	req := fmt.Sprintf("DELETE FROM usersP WHERE id=%s", user_id)
	output, err := send_sql_request(req)
	if err != nil {
		fmt.Fprintln(w, "Ошибка при удалении данных о пользователе")
		fmt.Fprint(w, output)
	}

	http.Get("auth.cgi")
	http.Redirect(w, r, "auth.cgi", http.StatusSeeOther)
}

func send_sql_request(req string) (string, error) {
	cmd := exec.Command("mysql", "-uu68871", "-p7773311", "-D", "u68871", "-e", req)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func check_JWT_token(r *http.Request, w http.ResponseWriter) (bool, string) {
	cookie, err := r.Cookie("jwt_token")
	if err != nil {
		fmt.Fprint(w, "Ошибка при получении куки с токеном ")
		fmt.Fprint(w, err)
		return false, ""
	}

	token_str := cookie.Value
	//fmt.Fprint(w, string(token_str)+"\n")
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(string(token_str), claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		fmt.Fprint(w, "Ошибка при проверке токена")
		return false, ""
	}

	return true, claims.Login
}
