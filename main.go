package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sort"
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
		usedLetters := ""

		letterPositions := make(map[rune][]int)
		for i, letter := range word {
			letterPositions[letter] = append(letterPositions[letter], i)
		}
		revealedPositions := make(map[int]bool)

		for i := 0; i < revealedLetters && i < len(word); i++ {
			var candidateLetters []rune
			for letter, positions := range letterPositions {
				if len(positions) > 1 && !revealedPositions[positions[0]] && !revealedPositions[positions[1]] {
					candidateLetters = append(candidateLetters, letter)
				}
			}

			var randIndex int
			var revealedLetter rune

			if len(candidateLetters) > 0 {

				revealedLetter = candidateLetters[rand.Intn(len(candidateLetters))]
				positions := letterPositions[revealedLetter]

				for _, pos := range positions {
					if !revealedPositions[pos] {
						randIndex = pos
						break
					}
				}
			} else {

				for {
					randIndex = rand.Intn(len(word))
					if !revealedPositions[randIndex] {
						revealedLetter = rune(word[randIndex])
						break
					}
				}
			}

			revealedPositions[randIndex] = true
			hiddenWord = hiddenWord[:randIndex*2] + string(revealedLetter) + hiddenWord[randIndex*2+1:]

			if !strings.Contains(usedLetters, string(revealedLetter)) {
				usedLetters += string(revealedLetter)
			}
		}

		setCookie(w, "username", username)
		setCookie(w, "difficulty", difficulty)
		setCookie(w, "word", word)
		setCookie(w, "hiddenWord", hiddenWord)
		setCookie(w, "attempts", attempts)
		setCookie(w, "usedLetters", usedLetters)

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
	usedLetters, _ := getCookie(r, "usedLetters")

	data := map[string]interface{}{
		"Username":    username,
		"Word":        word,
		"HiddenWord":  hiddenWord,
		"Attempts":    attempts,
		"Difficulty":  difficulty,
		"Image":       10 - attempts,
		"UsedLetters": strings.Join(strings.Split(usedLetters, ""), ", "),
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
	usedLetters, _ := getCookie(r, "usedLetters")
	image := ""
	guess := r.FormValue("guess")
	message := ""
	messageType := ""
	score := 0

	letterCount := strings.Count(word, guess)
	revealedCount := strings.Count(hiddenWord, guess)
	difficultyMultiplier := 1

	switch difficulty {
	case "medium":
		difficultyMultiplier = 2
	case "hard":
		difficultyMultiplier = 3
	}

	if strings.Contains(usedLetters, guess) {
		if letterCount > revealedCount {
			message = "Correct!"
			messageType = "success"
			image = strconv.Itoa(10 - attempts)

			score = 50 * letterCount * difficultyMultiplier * attempts

			for i, letter := range word {
				if string(letter) == guess &&
					hiddenWord[i*2:i*2+1] == "_" {
					hiddenWord = hiddenWord[:i*2] + guess + hiddenWord[i*2+1:]
				}
			}

			if strings.Replace(hiddenWord, " ", "", -1) == word {
				messageType = "success"
				message = "Congratulations, you won!"
				score = 500 * difficultyMultiplier * attempts

				score_board_entry := fmt.Sprintf("%s,%s,%s,%s,%d",
					username, word, attemptsStr, difficulty, score)
				f, err := os.OpenFile("./scoreboard.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				check(err)
				f.WriteString(score_board_entry + "\n")
				attempts = 0
			}
		} else {
			if revealedCount > 1 {
				message = "Toutes les occurrences de cette lettre ont été révélées !"
				messageType = "warning"
				image = strconv.Itoa(10 - attempts)
			} else {
				message = "Vous avez déjà utilisé cette lettre !"
				messageType = "warning"
				image = strconv.Itoa(10 - attempts)
			}
		}
	} else {
		usedLetters += guess
		setCookie(w, "usedLetters", usedLetters)

		if len(guess) > 1 {
			if guess == word {
				messageType = "success"
				hiddenWord = word
				image = "won"
				attempts = 0
				score = 500 * difficultyMultiplier * attempts
				message = "Congratulations, you won!"

				score_board_entry := fmt.Sprintf("%s,%s,%s,%s,%d",
					username, word, attemptsStr, difficulty, score)
				f, err := os.OpenFile("./scoreboard.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				check(err)
				f.WriteString(score_board_entry + "\n")
			} else {
				attempts -= 2
				message = "Incorrect!"
				messageType = "error"
				image = strconv.Itoa(10 - attempts)
			}
		} else {
			if strings.Contains(word, guess) {
				message = "Correct!"
				messageType = "success"
				image = strconv.Itoa(10 - attempts)
				score = 50 * letterCount * difficultyMultiplier * attempts

				for i, letter := range word {
					if string(letter) == guess {
						hiddenWord = hiddenWord[:i*2] + guess + hiddenWord[i*2+1:]
					}
				}
				if strings.Replace(hiddenWord, " ", "", -1) == word {
					messageType = "success"
					message = "Congratulations, you won!"
					image = "won"
					score = 500 * difficultyMultiplier * attempts

					score_board_entry := fmt.Sprintf("%s,%s,%s,%s,%d",
						username, word, attemptsStr, difficulty, score)
					f, err := os.OpenFile("./scoreboard.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
					check(err)
					f.WriteString(score_board_entry + "\n")
					attempts = 0
				}
			} else {
				attempts--
				if attempts == 0 {
					message = "You Lost! The word was: " + word
					messageType = "error"
					image = strconv.Itoa(10 - attempts)
				} else {
					message = "Incorrect!"
					messageType = "error"
					image = strconv.Itoa(10 - attempts)
				}
			}
		}
	}

	setCookie(w, "hiddenWord", hiddenWord)
	setCookie(w, "attempts", strconv.Itoa(attempts))

	data := map[string]interface{}{
		"Username":    username,
		"HiddenWord":  hiddenWord,
		"Attempts":    attempts,
		"Message":     message,
		"MessageType": messageType,
		"Difficulty":  difficulty,
		"UsedLetters": strings.Join(strings.Split(usedLetters, ""), ", "),
		"Image":       image,
		"Score":       score,
	}
	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "game.html", data)
}

type Score struct {
	Username   string
	Word       string
	Attempts   string
	Difficulty string
	Score      int
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
		if len(parts) == 5 {
			score, _ := strconv.Atoi(parts[4])
			scores = append(scores, Score{
				Username:   parts[0],
				Word:       parts[1],
				Attempts:   parts[2],
				Difficulty: parts[3],
				Score:      score,
			})
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	data := map[string]interface{}{
		"Scores": scores,
	}

	w.Header().Set("Content-Type", "text/html")
	templates.ExecuteTemplate(w, "scores.html", data)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/game", gameHandler)
	http.HandleFunc("/game/submit", submitHandler)
	http.HandleFunc("/scoreboard", scoreHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	fmt.Println("Starting server on http://localhost:8080...")
	http.ListenAndServe(":8080", nil)
}
