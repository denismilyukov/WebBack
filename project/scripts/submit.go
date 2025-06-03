package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"net/http/cgi"
	"os/exec"
	"strings"
)

type FormData struct {
	Fio      string
	Phone    string
	Email    string
	Org      string
	Bio      string
	Contract string
}

type User struct {
	Fio      string
	Login    string
	Password string
}

func main() {
	var err error
	err = cgi.Serve(http.HandlerFunc(handler))
	if err != nil {
		fmt.Println("Content-type: text/plain\n")
		fmt.Println("Failed to serve CGI request")
	}
}

func send_sql_request(req string) (string, error) {
	cmd := exec.Command("mysql", "-uu68871", "-p7773311", "-D", "u68871", "-e", req)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	formData := FormData{
		Fio:      r.FormValue("fio"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		Org:      r.FormValue("org"),
		Bio:      r.FormValue("message"),
		Contract: r.FormValue("policy"),
	}
	flag := 0
	if formData.Contract == "on" {
		flag = 1
	}

	// is_valid, val_answer := validate_data(formData)
	// if !is_valid {
	// 	//fmt.Fprint(w, val_answer)
	// 	save_form_errors(w, val_answer)
	// 	save_form_data(w, formData)
	// 	// c, er := r.Cookie("form_data")
	// 	// if er == nil {
	// 	// 	fmt.Fprint(w, c.Value)
	// 	// } else {
	// 	// 	fmt.Fprint(w, er)
	// 	// }
	// 	http.Redirect(w, r, "submit.cgi", http.StatusSeeOther)
	// 	return
	// }

	req := fmt.Sprintf("INSERT INTO usersP (fio, phone, mail, org, bio, contract) VALUES ('%s', '%s', '%s', '%s', '%s', %d);",
		formData.Fio, formData.Phone, formData.Email, formData.Org, formData.Bio, flag)
	output, err := send_sql_request(req)
	if err != nil {
		fmt.Fprintln(w, "Ошибка при выполнении sql-запроса при записи данных. Возможно вы ввели недопустимые символы в биографии")
		fmt.Fprintln(w, output)
		fmt.Fprintln(w, formData)
		return
	}

	login := generate_login(6)
	password := generate_pass(10)
	hashed_pass := sha256.Sum256([]byte(password))

	req = "SELECT MAX(id) FROM usersP;"
	output, _ = send_sql_request(req)
	last_user_id := strings.Split(string(output), "\n")[2]

	req = fmt.Sprintf("INSERT INTO sensitivesP (user_id, login, pass) VALUES (%s, '%s', '%s');", last_user_id, login, hex.EncodeToString(hashed_pass[:]))
	_, err = send_sql_request(req)
	if err != nil {
		fmt.Fprint(w, "Ошибка выполнения sql-запроса при записи чувствительных данных")
		return
	}

	// fmt.Fprint(w, formData.Fio+" , ваши данные успешно сохранены\n")
	// fmt.Fprintf(w, "Запомните и никому не сообщайте!\nВаш логин: %s\nВаш пароль: %s", login, password)
	tmpl, err := template.ParseFiles("success_reg.html")
	user := User{
		Fio:      formData.Fio,
		Login:    login,
		Password: password,
	}
	if err != nil {
		fmt.Fprint(w, "Ошибка парсинга шаблона")
	}
	tmpl.ExecuteTemplate(w, "success_reg.html", user)
	return
}

func generate_login(length int) string {
	digits := "0123456789"
	res := make([]byte, length-1)

	for i := range res {
		res[i] = digits[rand.Intn(10)]
	}

	return "u" + string(res)
}

func generate_pass(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	res := make([]byte, length)

	for i := range res {
		res[i] = chars[rand.Intn(len(chars))]
	}
	return string(res)
}
