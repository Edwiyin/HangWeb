package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func check(e error) {
	if e != nil {
		panic(e)
	}
}
func init() {
	rand.Seed(time.Now().UnixNano())
}

func getRandomWord(difficulty string) (string, error) {

	filePath := difficulty + ".txt"

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	words := strings.Split(strings.TrimSpace(string(content)), "\n")
	return words[rand.Intn(len(words))], nil
}

func setCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	})
}

func getCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		difficulty := r.FormValue("difficulty")
		attempts := ""
		revealedLetters := 0
		switch difficulty {
		case "easy":
			attempts = "10"
			revealedLetters = 2
		case "medium":
			attempts = "7"
			revealedLetters = 1
		case "hard":
			attempts = "5"
			revealedLetters = 0
		}

		word, err := getRandomWord(difficulty)
		if err != nil {
			http.Error(w, "Failed to load word list", http.StatusInternalServerError)
			return
		}

		hiddenWord := strings.Repeat("_ ", len(word)-1)
		for i := 0; i < revealedLetters; i++ {
			randIndex := -1
			oldIndex := -1
			if randIndex != -1 {
				oldIndex = randIndex
			}
			randIndex = rand.Intn(len(word))

			for randIndex == oldIndex {
				randIndex = rand.Intn(len(word))
			}
			hiddenWord = hiddenWord[:randIndex*2] + string(word[randIndex]) + hiddenWord[randIndex*2+1:]
		}

		setCookie(w, "username", username)
		setCookie(w, "difficulty", difficulty)
		setCookie(w, "word", word)
		setCookie(w, "hiddenWord", hiddenWord)
		setCookie(w, "attempts", attempts)

		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "index.html", nil)
}

func gameHandler(w http.ResponseWriter, r *http.Request) {

	username, _ := getCookie(r, "username")
	word, _ := getCookie(r, "word")
	hiddenWord, _ := getCookie(r, "hiddenWord")
	attemptsStr, _ := getCookie(r, "attempts")
	attempts, _ := strconv.Atoi(attemptsStr)
	difficulty, _ := getCookie(r, "difficulty")

	data := map[string]interface{}{
		"Username":   username,
		"Word":       word,
		"HiddenWord": hiddenWord,
		"Attempts":   attempts,
		"Difficulty": difficulty,
		"Image":      10 - attempts,
	}

	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "game.html", data)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {

	username, _ := getCookie(r, "username")
	word, _ := getCookie(r, "word")
	hiddenWord, _ := getCookie(r, "hiddenWord")
	attemptsStr, _ := getCookie(r, "attempts")
	attempts, _ := strconv.Atoi(attemptsStr)
	difficulty, _ := getCookie(r, "difficulty")

	guess := r.FormValue("guess")
	message := ""
	messageType := ""

	if strings.Contains(word, guess) {
		message = "Correct!"
		messageType = "success"
		for i, letter := range word {
			if string(letter) == guess {
				hiddenWord = hiddenWord[:i*2] + guess + hiddenWord[i*2+1:]
			}
		}
	} else {
		attempts--
		message = "Incorrect!"
		messageType = "error"
	}

	setCookie(w, "hiddenWord", hiddenWord)
	setCookie(w, "attempts", strconv.Itoa(attempts))

	if !strings.Contains(hiddenWord, "_") {
		http.Redirect(w, r, "/end?result=win", http.StatusSeeOther)
		return
	}
	if attempts < 0 {
		http.Redirect(w, r, "/end?result=lose", http.StatusSeeOther)
		return
	}

	image := 10 - attempts

	data := map[string]interface{}{
		"Username":    username,
		"HiddenWord":  hiddenWord,
		"Attempts":    attempts,
		"Message":     message,
		"MessageType": messageType,
		"Difficulty":  difficulty,
		"Image":       image,
	}
	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "game.html", data)
}

type Score struct {
	Username string
	Word     string
	Attempts string
}

func scoreHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.ReadFile("scoreboard.txt")
	if err != nil {
		http.Error(w, "Failed to read scoreboard", http.StatusInternalServerError)
		return
	}

	var scores []Score
	lines := strings.Split(strings.TrimSpace(string(file)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) == 3 {
			scores = append(scores, Score{
				Username: parts[0],
				Word:     parts[1],
				Attempts: parts[2],
			})
		}
	}

	data := map[string]interface{}{
		"Scores": scores,
	}

	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "scores.html", data)
}
func endHandler(w http.ResponseWriter, r *http.Request) {
	result := r.URL.Query().Get("result")
	message := ""
	if result == "win" {
		message = "Congratulations, you won!"
		word, _ := getCookie(r, "word")
		username, _ := getCookie(r, "username")
		attempts, _ := getCookie(r, "attempts")
		score_board_entry := username + "," + word + "," + attempts
		f, err := os.OpenFile("./scoreboard.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		check(err)
		f.WriteString(score_board_entry + "\n")
	} else {
		word, _ := getCookie(r, "word")
		message = "Sorry, you lost! The word was: " + word
	}

	http.SetCookie(w, &http.Cookie{Name: "username", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "difficulty", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "word", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "hiddenWord", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "attempts", Value: "", Path: "/", MaxAge: -1})

	data := map[string]interface{}{
		"Message": message,
	}

	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "end.html", data)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/game/submit", submitHandler)
	http.HandleFunc("/end", endHandler)
	http.HandleFunc("/scoreboard", scoreHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	fmt.Println("Starting server on http://localhost:8080...")
	http.ListenAndServe(":8080", nil)
}
