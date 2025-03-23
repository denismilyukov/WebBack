package main

import (
	"fmt"
	"net/http"
	"net/http/cgi"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type FormData struct {
	Fio      string
	Phone    string
	Email    string
	Dob      string
	Gender   string
	Bio      string
	Langs    []string
	Contract bool
}

func main() {
	var err error
	err = cgi.Serve(http.HandlerFunc(handler))
	if err != nil {
		fmt.Println("Content-type: text/plain\n")
		fmt.Println("Failed to serve CGI request")
	}
}

func send_sql_request(req string) ([]byte, error) {
	cmd := exec.Command("mysql", "-uu68871", "-p7773311", "-D", "u68871", "-e", req)
	output, err := cmd.CombinedOutput()
	return output, err
}

func validate_data(formData FormData) (bool, string) {

	if formData.Fio == "" {
		return false, "Поле 'ФИО' обязательно для заполнения"
	}
	re := regexp.MustCompile(`^([a-zA-z]+\s){2}[a-zA-z]+$`)
	if !re.MatchString(formData.Fio) || len(formData.Fio) > 150 {
		return false, "Введите ФИО корректно, оно должно содержать только латинские буквы, а длина не должна превышать 150 символов"
	}

	if formData.Email == "" {
		return false, "Поле 'Почта' обязательно для заполнения"
	}
	re = regexp.MustCompile(`^[\w\.-_]+@[a-zA-Z]+\.[a-zA-z]+$`)
	if !re.MatchString(formData.Email) {
		return false, "Введите адрес почты корректно, она должна соответствовать форме adress@mail.domen"
	}

	if formData.Phone == "" {
		return false, "Поле 'Телефон' обязательно для заполнения"
	}
	re = regexp.MustCompile(`^\+\d{11}$`)
	if !re.MatchString(formData.Phone) {
		return false, "Введите номер телефона корректно, он должен начинаться с + и после этого содержать 11 цифр"
	}

	if len(formData.Langs) == 0 {
		return false, "Выбор любимых языков программирования обязателен! Выберите хотя бы Pascal"
	}

	re = regexp.MustCompile(`^\d{4}(-\d{2}){2}$`)
	if !re.MatchString(formData.Dob) {
		return false, "Поле ввода даты обязательно для заполнения"
	}

	if formData.Bio == "" {
		return false, "Поле ввода биографии обязательно для заполнения"
	}

	if formData.Contract == false {
		return false, "Ознакомление с контрактом обязательно"
	}

	return true, fmt.Sprintf("'%s', Ваши данные успешно сохранены!", formData.Fio)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, "index.html")
		return
	}
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
		Dob:      r.FormValue("date"),
		Gender:   r.FormValue("gender"),
		Bio:      r.FormValue("message"),
		Contract: r.FormValue("policy") == "on",
		Langs:    r.Form["langs"],
	}
	flag := 0
	if formData.Contract == true {
		flag = 1
	}

	is_valid, val_answer := validate_data(formData)
	if !is_valid {
		fmt.Fprint(w, val_answer)
		return
	}

	req := fmt.Sprintf("INSERT INTO users (fio, gender, phone, mail, date, bio, contract) VALUES ('%s', '%s', '%s', '%s', '%s', '%s', %d);", formData.Fio,
		formData.Gender, formData.Phone, formData.Email, formData.Dob, formData.Bio, flag)
	output, err := send_sql_request(req)
	req = "SELECT MAX(id) FROM users;"
	output, err = send_sql_request(req)
	last_user_id, err := strconv.Atoi(strings.Split(string(output), "\n")[2])
	for _, lang_id := range formData.Langs {
		lang, _ := strconv.Atoi(lang_id)
		req = fmt.Sprintf("INSERT INTO languages_on_user (user_id, lang_id) VALUES (%d, %d);", last_user_id, lang)
		output, err = send_sql_request(req)
	}
	//w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, val_answer)
}
