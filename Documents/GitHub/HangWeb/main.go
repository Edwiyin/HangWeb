// main.go
package main

import (
    "html/template"
    "net/http"
    "sync"
)

type GameState struct {
    Username     string
    Difficulty   string
    Word         string
    HiddenWord   string
    UsedLetters  []string
    Lives        int
    Message      string
    GameOver     bool
    Won          bool
    mu           sync.Mutex
}

type Score struct {
    Username   string
    Difficulty string
    Word       string
    Won        bool
    LivesLeft  int
}

type GameManager struct {
    Sessions map[string]*GameState
    Scores   []Score
    mu       sync.Mutex
}

var (
    gameManager = &GameManager{
        Sessions: make(map[string]*GameState),
        Scores:   make([]Score, 0),
    }
    templates = template.Must(template.ParseGlob("templates/*.html"))
)

func main() {
   
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    
    http.HandleFunc("/", handleHome)
    http.HandleFunc("/game", handleGame)
    http.HandleFunc("/play", handlePlay)
    http.HandleFunc("/scores", handleScores)
    http.HandleFunc("/gameover", handleGameOver)
   
    println("Serveur démarré sur http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}

func handleHome(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        templates.ExecuteTemplate(w, "home.html", nil)
        return
    }
    
    if r.Method == http.MethodPost {
        username := r.FormValue("username")
        difficulty := r.FormValue("difficulty")
      
        gameState := &GameState{
            Username:   username,
            Difficulty: difficulty,
            Lives:      6,
        }
       
        gameManager.mu.Lock()
        gameManager.Sessions[username] = gameState
        gameManager.mu.Unlock()
        
        http.Redirect(w, r, "/game", http.StatusSeeOther)
    }
}

func handleGame(w http.ResponseWriter, r *http.Request) {
    
}

func handlePlay(w http.ResponseWriter, r *http.Request) {
    
}

func handleScores(w http.ResponseWriter, r *http.Request) {
  
}

func handleGameOver(w http.ResponseWriter, r *http.Request) {
   
}