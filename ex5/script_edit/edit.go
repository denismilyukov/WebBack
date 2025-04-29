package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/http/cgi"
	"net/url"
	"os/exec"
	"regexp"
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

func validate_data(formData FormData) (bool, FormErrors) {
	data := FormErrors{}
	flag := true
	re := regexp.MustCompile(`^([a-zA-z]+\s){2}[a-zA-z]+$`)

	if formData.Fio == "" {
		flag = false
		data.Fio = "Поле 'ФИО' обязательно для заполнения"
	} else if !re.MatchString(formData.Fio) || len(formData.Fio) > 150 {
		flag = false
		data.Fio = "Введите ФИО корректно, оно должно содержать только латинские буквы, а длина не должна превышать 150 символов"
	}

	re = regexp.MustCompile(`^[\w\.-_]+@[a-zA-Z]+\.[a-zA-z]+$`)
	if formData.Email == "" {
		flag = false
		data.Email = "Поле 'Почта' обязательно для заполнения"
	} else if !re.MatchString(formData.Email) {
		flag = false
		data.Email = "Введите адрес почты корректно, она должна соответствовать форме adress@mail.domen"
	}

	re = regexp.MustCompile(`^\+\d{11}$`)
	if formData.Phone == "" {
		flag = false
		data.Phone = "Поле 'Телефон' обязательно для заполнения"
	} else if !re.MatchString(formData.Phone) {
		flag = false
		data.Phone = "Введите номер телефона корректно, он должен начинаться с + и после этого содержать 11 цифр"
	}

	if len(formData.Langs) == 0 {
		flag = false
		data.Langs = "Выбор любимых языков программирования обязателен! Выберите хотя бы Pascal"
	}

	re = regexp.MustCompile(`^\d{4}(-\d{2}){2}$`)
	if !re.MatchString(formData.Dob) {
		flag = false
		data.Dob = "Поле ввода даты обязательно для заполнения"
	}

	if formData.Bio == "" {
		flag = false
		data.Bio = "Поле ввода биографии обязательно для заполнения"
	}

	if formData.Contract != "on" {
		flag = false
		data.Contract = "Ознакомление с контрактом обязательно"
	}

	return flag, data
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

func handler_form(w http.ResponseWriter, r *http.Request) {
	data := PageData{}
	//http.ServeFile(w, r, "index.html")
	//tmpl.Funcs(template.FuncMap{"not_empty": not_empty})
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
	//fmt.Fprint(w, "Обращение к cookie для получения информации об ошибках")
	cookie, err := r.Cookie("form_data")
	if err == nil {
		decoded_data, _ := url.QueryUnescape(cookie.Value)
		info := strings.Split(decoded_data, "/")
		//fmt.Fprint(w, info)
		data.Data.Fio = info[0]
		data.Data.Email = info[1]
		data.Data.Phone = info[2]
		if len(info[3]) > 0 {
			data.Data.Langs = strings.Split(info[3], ",")
		} else {
			data.Data.Langs = make([]string, 0)
		}
		data.Data.Dob = info[4]
		data.Data.Gender = info[5]
		data.Data.Bio = info[6]
		data.Data.Contract = info[7]
		data.Data.Id = info[8]
	} else if err == http.ErrNoCookie {
		//http.Redirect(w, r, "index.html", http.StatusSeeOther)
		fmt.Println("Cookie с именем form_data не существует")
		return
	} else {
		fmt.Println("Неизвестная ошибка при получении Cookie")
	}

	cookie, err = r.Cookie("form_errors")
	if err == nil {
		decoded_data, _ := url.QueryUnescape(cookie.Value)
		info := strings.Split(decoded_data, "/")
		data.Errors.Fio = info[0]
		data.Errors.Email = info[1]
		data.Errors.Phone = info[2]
		data.Errors.Langs = info[3]
		data.Errors.Dob = info[4]
		data.Errors.Bio = info[5]
		data.Errors.Contract = info[6]
		data.IsError = has_errors(data.Errors)
		http.SetCookie(w, &http.Cookie{
			Name:   "form_errors",
			Value:  "",
			MaxAge: -1,
		})
	} //else if err == http.ErrNoCookie {
	//http.Redirect(w, r, "index.html", http.StatusSeeOther)
	//fmt.Fprint(w, "Cookie с таким именем не найдена")
	//return
	//} else {
	//	fmt.Fprint(w, "Ошибка Cookie")
	//}

	//http.Redirect(w, r, "index.html", http.StatusSeeOther)
	//fmt.Println(data)
	tmpl.ExecuteTemplate(w, "edit.html", data)
}

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		handler_form(w, r)
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
		Contract: r.FormValue("policy"),
		Langs:    r.Form["langs"],
		Id:       r.FormValue("id"),
	}
	flag := 0
	if formData.Contract == "on" {
		flag = 1
	}

	is_valid, val_answer := validate_data(formData)
	if !is_valid {
		//fmt.Fprint(w, val_answer)
		save_form_errors(w, val_answer)
		save_form_data(w, formData)
		// c, er := r.Cookie("form_data")
		// if er == nil {
		// 	fmt.Fprint(w, c.Value)
		// } else {
		// 	fmt.Fprint(w, er)
		// }
		http.Redirect(w, r, "edit.cgi", http.StatusSeeOther)
		return
	}

	req := fmt.Sprintf("UPDATE users SET fio='%s', gender='%s', phone='%s', mail='%s', date='%s', bio='%s', contract=%d WHERE id=%s;",
		formData.Fio, formData.Gender, formData.Phone, formData.Email, formData.Dob, formData.Bio, flag, formData.Id)
	_, err = send_sql_request(req)
	if err != nil {
		fmt.Fprint(w, "Ошибка выполнения sql-запроса при обновлении данных")
		fmt.Fprint(w, err)
		return
	}

	old_langs := get_old_langs(formData.Id)
	if has_new_langs(old_langs, formData.Langs) {
		req = fmt.Sprintf("DELETE FROM languages_on_user WHERE user_id=%s;", formData.Id)
		_, err = send_sql_request(req)
		if err != nil {
			fmt.Fprint(w, "Ошибка выполнения sql-запроса при обновлении списка ЯП(1)")
		}

		for _, lang_id := range formData.Langs {
			req = fmt.Sprintf("INSERT INTO languages_on_user (user_id, lang_id) VALUES (%s, %s);", formData.Id, lang_id)
			_, err := send_sql_request(req)
			if err != nil {
				fmt.Fprint(w, "Ошибка выполнения sql-запроса при обновлении списка ЯП(2)")
			}
		}
	}

	fmt.Fprint(w, formData.Fio+", ваши данные успешно обновлены!")
}

func get_old_langs(user_id string) []string {
	req := fmt.Sprintf("SELECT lang_id FROM languages_on_user WHERE user_id=%s;", user_id)
	output, _ := send_sql_request(req)
	var langs []string
	for _, word := range strings.Split(output, "\n")[2:] {
		if word != "" {
			langs = append(langs, word)
		}
	}
	return langs
}

func has_new_langs(old_langs []string, new_langs []string) bool {
	if len(old_langs) != len(new_langs) {
		return true
	}

	for i, _ := range old_langs {
		if old_langs[i] != new_langs[i] {
			return true
		}
	}

	return false
}

func save_form_errors(w http.ResponseWriter, errors FormErrors) {
	serialized_data := fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s", errors.Fio, errors.Email, errors.Phone, errors.Langs,
		errors.Dob, errors.Bio, errors.Contract)
	encoded_data := url.QueryEscape(serialized_data)
	http.SetCookie(w, &http.Cookie{
		Name:     "form_errors",
		Value:    encoded_data,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	//fmt.Fprint(w, "Данные об ошибках сохранены в куки \n"+serialized_data)
}

func save_form_data(w http.ResponseWriter, formData FormData) {
	langs := strings.Join(formData.Langs, ",")
	serialized_data := fmt.Sprintf("%s/%s/%s/%s/%s/%s/%s/%s/%s", formData.Fio, formData.Email, formData.Phone, langs,
		formData.Dob, formData.Gender, formData.Bio, formData.Contract, formData.Id)
	encoded_data := url.QueryEscape(serialized_data)
	http.SetCookie(w, &http.Cookie{
		Name:     "form_data",
		Value:    encoded_data,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	//fmt.Fprint(w, "Данные сохранены в куки \n"+serialized_data)
}

func has_errors(errors FormErrors) bool {
	return len(errors.Fio) > 0 || len(errors.Email) > 0 || len(errors.Phone) > 0 || len(errors.Langs) > 0 || len(errors.Dob) > 0 || len(errors.Bio) > 0 || len(errors.Contract) > 0
}
