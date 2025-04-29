package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"net/http/cgi"
	"os/exec"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("sVnOd8XxdyH2M2q0X0SphbaAmK81cKM2")

type Claims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

type User struct {
	Login    string
	Password string
}

type FormData struct {
	Fio      string
	Phone    string
	Email    string
	Dob      string
	Gender   string
	Bio      string
	Langs    []string
	Contract string
	Id       string
}

type FormErrors struct {
	Fio      string
	Phone    string
	Email    string
	Dob      string
	Bio      string
	Langs    string
	Contract string
}

type PageData struct {
	Data    FormData
	Errors  FormErrors
	IsError bool
}

func main() {
	err := cgi.Serve(http.HandlerFunc(handler))
	if err != nil {
		fmt.Println("Failed to serve CGI request")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handler_get(w, r)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	user_data := User{
		Login:    r.FormValue("login"),
		Password: r.FormValue("pass"),
	}

	req := fmt.Sprintf("SELECT pass FROM sensitives WHERE login='%s';", user_data.Login)
	output, _ := send_sql_request(req)
	//fmt.Fprint(w, output)
	output_strokes := strings.Split(output, "\n")
	// fmt.Fprint(w, output_strokes)
	// fmt.Fprint(w, len(output_strokes))
	var pass_from_db string
	if len(output_strokes) > 2 {
		pass_from_db = output_strokes[2]
		//fmt.Fprint(w, pass_from_db)
		if check_password(pass_from_db, user_data.Password) {
			//fmt.Fprint(w, "Вы верно ввели пароль!")
			token_str := generate_JWT_token(user_data.Login, w)
			//fmt.Fprint(w, token_str+"\n")
			save_JWT_token(token_str, w)
			// if check_JWT_token(r, w) {
			// 	build_form(user_data.Login, w)
			// }
			http.Redirect(w, r, "auth.cgi", http.StatusSeeOther)
			return
		} else {
			fmt.Fprint(w, "Вы неверно ввели пароль(")
		}

	} else {
		fmt.Fprint(w, "Вы неверно ввели логин")
	}
}

func handler_get(w http.ResponseWriter, r *http.Request) {
	flag, login := check_JWT_token(r, w)
	if flag {
		//fmt.Fprint(w, login)
		build_form(login, w)
	}
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

func save_JWT_token(token string, w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:    "jwt_token",
		Value:   token,
		Expires: time.Now().Add(25 * time.Second),
		// HttpOnly: true,
		// SameSite: http.SameSiteStrictMode,
		// Secure:   false,
		// Path:     "/",
	})
}

func generate_JWT_token(login string, w http.ResponseWriter) string {
	exp_time := time.Now().Add(25 * time.Second)
	claims := &Claims{
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp_time),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token_str, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
	}
	return token_str
}

func build_form(login string, w http.ResponseWriter) {
	formData := get_data_on_login(login)
	data := PageData{
		Data:    formData,
		IsError: false,
	}

	tmpl := template.New("").Funcs(template.FuncMap{
		"not_empty":   not_empty,
		"is_selected": is_selected,
		"is_checked":  is_checked,
		"is_male":     is_male,
	})
	tmpl, err := tmpl.ParseFiles("edit.html")
	if err != nil {
		fmt.Fprint(w, "Ошибка парсинга шаблона")
	}

	tmpl.ExecuteTemplate(w, "edit.html", data)
}

func get_data_on_login(login string) FormData {
	req := fmt.Sprintf("SELECT user_id FROM sensitives WHERE login='%s';", login)
	output, _ := send_sql_request(req)
	user_id := strings.Split(output, "\n")[2]

	req = fmt.Sprintf("SELECT * FROM users WHERE id=%s", user_id)
	output, _ = send_sql_request(req)
	//fmt.Fprint(w, output)

	var data []string
	output = strings.Split(output, "\n")[2]
	for _, word := range strings.Split(output, "\t") {
		data = append(data, word)
		//fmt.Fprintf(w, "%d: "+word+"\n", i)
	}

	req = fmt.Sprintf("SELECT lang_id FROM languages_on_user WHERE user_id=%s", user_id)
	output, _ = send_sql_request(req)
	var langs []string
	for _, word := range strings.Split(output, "\n")[2:] {
		if word != "" {
			langs = append(langs, word)
		}
	}
	//fmt.Fprint(w, output)

	formData := FormData{
		Id:       data[0],
		Fio:      data[1],
		Gender:   data[2],
		Phone:    data[3],
		Email:    data[4],
		Dob:      data[5],
		Bio:      data[6],
		Contract: "on",
		Langs:    langs,
	}

	return formData
	//fmt.Fprint(w, formData)
}

func check_password(hashed_pass string, form_pass string) bool {
	hashed_form_pass := sha256.Sum256([]byte(form_pass))
	return hex.EncodeToString(hashed_form_pass[:]) == hashed_pass
}

func send_sql_request(req string) (string, error) {
	cmd := exec.Command("mysql", "-uu68871", "-p7773311", "-D", "u68871", "-e", req)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func not_empty(s string) bool {
	return len(strings.TrimSpace(s)) > 0
}

func is_selected(lang string, langs []string) bool {
	for _, selected_lang := range langs {
		if lang == selected_lang {
			return true
		}
	}
	return false
}

func is_checked(s string) bool {
	return s == "on"
}

func is_male(s string) bool {
	return s == "male"
}
