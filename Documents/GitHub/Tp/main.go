package main

import (
	"html/template"
	"net/http"
	"regexp"
	"sync"
)

type Student struct {
	FirstName string
	LastName  string
	Age       int
	Gender    string
}

type Class struct {
	Name         string
	Field        string
	Level        string
	StudentCount int
	StudentsList []Student
}

type ViewData struct {
	Count   int
	Message string
}

type UserData struct {
	LastName      string
	FirstName     string
	BirthDate     string
	Gender        string
	ErrorMessage  string
	IsFormSuccess bool
}

var (
	viewCount int
	userData  UserData
	mutex     sync.Mutex

	letterOnlyRegex = regexp.MustCompile("^[a-zA-ZÀ-ÿ\\s-]+$")
)

func main() {

	http.HandleFunc("/promo", promoHandler)
	http.HandleFunc("/change", changeHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/user/form", userFormHandler)
	http.HandleFunc("/user/treatment", userTreatmentHandler)
	http.HandleFunc("/user/display", userDisplayHandler)

	http.ListenAndServe(":8080", nil)
}

func promoHandler(w http.ResponseWriter, r *http.Request) {
	class := Class{
		Name:         "B1 Informatique",
		Field:        "Informatique",
		Level:        "Bachelor 1",
		StudentCount: 3,
		StudentsList: []Student{
			{FirstName: "Jean", LastName: "Dupont", Age: 20, Gender: "M"},
			{FirstName: "Marie", LastName: "Martin", Age: 19, Gender: "F"},
			{FirstName: "Pierre", LastName: "Gustav", Age: 21, Gender: "M"},
		},
	}

	tmpl, err := template.ParseFiles("promo.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, class)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func changeHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	viewCount++
	currentCount := viewCount
	mutex.Unlock()

	var message string
	if currentCount%2 == 0 {
		message = "Le nombre de vues est pair"
	} else {
		message = "Le nombre de vues est impair"
	}

	data := ViewData{
		Count:   currentCount,
		Message: message,
	}

	tmpl, err := template.ParseFiles("change.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func userFormHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/user_form.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func validateUserData(data *UserData) bool {
	
	if len(data.LastName) < 1 || len(data.LastName) > 32 || !letterOnlyRegex.MatchString(data.LastName) {
		return false
	}


	if len(data.FirstName) < 1 || len(data.FirstName) > 32 || !letterOnlyRegex.MatchString(data.FirstName) {
		return false
	}

	if data.Gender != "masculin" && data.Gender != "féminin" && data.Gender != "autre" {
		return false
	}

	if data.BirthDate == "" {
		return false
	}

	return true
}

func userTreatmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/user/form", http.StatusSeeOther)
		return
	}

	mutex.Lock()
	userData = UserData{
		LastName:  r.FormValue("lastname"),
		FirstName: r.FormValue("firstname"),
		BirthDate: r.FormValue("birthdate"),
		Gender:    r.FormValue("gender"),
	}
	mutex.Unlock()

	if validateUserData(&userData) {
		userData.IsFormSuccess = true
		http.Redirect(w, r, "/user/display", http.StatusSeeOther)
	} else {
		userData.ErrorMessage = "Données invalides. Veuillez vérifier vos informations."
		http.Redirect(w, r, "/user/form", http.StatusSeeOther)
	}
}

func userDisplayHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	data := userData
	mutex.Unlock()

	tmpl, err := template.ParseFiles("templates/user_display.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !data.IsFormSuccess {
		data.ErrorMessage = "Veuillez d'abord renseigner vos informations personnelles."
	}

	tmpl.Execute(w, data)
}