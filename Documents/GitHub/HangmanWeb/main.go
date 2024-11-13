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
	rand.Seed(time.Now().UnixNano()) // Seed random generator
}

// Helper function to get a random word based on difficulty
func getRandomWord(difficulty string) (string, error) {
	// Map difficulty to file path
	filePath := difficulty + ".txt" // easy.txt, medium.txt, or hard.txt

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Split content into words and select a random word
	words := strings.Split(strings.TrimSpace(string(content)), "\n")
	return words[rand.Intn(len(words))], nil
}

// Helper to set cookies
func setCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:  name,
		Value: value,
		Path:  "/",
	})
}

// Helper to get cookies
func getCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// Handler for the index page (start page)
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		difficulty := r.FormValue("difficulty")

		// Get a random word based on difficulty
		word, err := getRandomWord(difficulty)
		if err != nil {
			http.Error(w, "Failed to load word list", http.StatusInternalServerError)
			return
		}
		// Add code for difficulty revealing letters and number of tries
		hiddenWord := strings.Repeat("_ ", len(word)-1)
		attempts := "7"

		// Set cookies to persist game state
		setCookie(w, "username", username)
		setCookie(w, "difficulty", difficulty)
		setCookie(w, "word", word)
		setCookie(w, "hiddenWord", hiddenWord)
		setCookie(w, "attempts", attempts)

		// Redirect to the game page
		http.Redirect(w, r, "/game", http.StatusSeeOther)
		return
	}

	// Render the index template
	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "index.html", nil)
}

// Handler for the game page
func gameHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve game state from cookies
	username, _ := getCookie(r, "username")
	word, _ := getCookie(r, "word")
	hiddenWord, _ := getCookie(r, "hiddenWord")
	attemptsStr, _ := getCookie(r, "attempts")
	attempts, _ := strconv.Atoi(attemptsStr)

	// Pass data to the template for display
	data := map[string]interface{}{
		"Username":   username,
		"Word":       word,
		"HiddenWord": hiddenWord,
		"Attempts":   attempts,
	}

	// Render game page
	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "game.html", data)
}

// Handler to process guesses
func submitHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve game data from cookies
	word, _ := getCookie(r, "word")
	hiddenWord, _ := getCookie(r, "hiddenWord")
	attemptsStr, _ := getCookie(r, "attempts")
	attempts, _ := strconv.Atoi(attemptsStr)

	guess := r.FormValue("guess")
	message := ""

	// Check if guess is in the word
	if strings.Contains(word, guess) {
		message = "Correct!"
		for i, letter := range word {
			if string(letter) == guess {
				hiddenWord = hiddenWord[:i*2] + guess + hiddenWord[i*2+1:]
			}
		}
	} else {
		attempts--
		message = "Incorrect!"
	}

	// Update cookies with the new game state
	setCookie(w, "hiddenWord", hiddenWord)
	setCookie(w, "attempts", strconv.Itoa(attempts))

	// Check if the game is won or lost
	if !strings.Contains(hiddenWord, "_") {
		http.Redirect(w, r, "/end?result=win", http.StatusSeeOther)
		return
	}
	if attempts <= 0 {
		http.Redirect(w, r, "/end?result=lose", http.StatusSeeOther)
		return
	}

	// Reload game page with the updated state and message
	data := map[string]interface{}{
		"HiddenWord": hiddenWord,
		"Attempts":   attempts,
		"Message":    message,
	}
	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "game.html", data)
}

// Handler for the end game page
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

	// Clear game cookies
	http.SetCookie(w, &http.Cookie{Name: "username", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "difficulty", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "word", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "hiddenWord", Value: "", Path: "/", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "attempts", Value: "", Path: "/", MaxAge: -1})

	// Render the end game page with the message
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

	fmt.Println("Starting server on http://localhost:8080...")
	http.ListenAndServe(":8080", nil)
}
