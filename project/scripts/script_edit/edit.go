package main

import (
	"fmt"
	"net/http"
	"net/http/cgi"
	"os/exec"
)

type FormData struct {
	Fio      string
	Phone    string
	Email    string
	Org      string
	Bio      string
	Contract string
	Id       string
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
		Id:       r.FormValue("id"),
	}
	flag := 0
	if formData.Contract == "on" {
		flag = 1
	}

	req := fmt.Sprintf("UPDATE usersP SET fio='%s', phone='%s', mail='%s', org='%s', bio='%s', contract=%d WHERE id=%s;",
		formData.Fio, formData.Phone, formData.Email, formData.Org, formData.Bio, flag, formData.Id)
	_, err = send_sql_request(req)
	if err != nil {
		fmt.Fprint(w, "Ошибка выполнения sql-запроса при обновлении данных")
		fmt.Fprint(w, err)
		fmt.Fprint(w, formData)
		return
	}

	http.Redirect(w, r, "auth.cgi", http.StatusSeeOther)
}
