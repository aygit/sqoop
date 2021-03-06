package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/solo-io/gloo/pkg/log"
	"github.com/solo-io/sqoop/examples/starwars/imported/starwars"
)

var baseResolvers = starwars.NewResolver()

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request path
		log.Printf("%v", r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// our main function
func New() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/api/hero", GetHero).Methods("GET")
	router.HandleFunc("/api/character", GetCharacter).Methods("GET")
	// needs to be POST because there's a body
	router.HandleFunc("/api/characters", GetCharacters).Methods("POST")

	router.Use(loggingMiddleware)

	return router
}

func GetHero(w http.ResponseWriter, r *http.Request) {
	hero, err := baseResolvers.Query_hero(r.Context(), "")
	if err != nil {
		log.Printf("request failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(hero); err != nil {
		log.Printf("encoding failed: %v", err)
	}
}

func GetCharacter(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get("x-id")
	friends, err := baseResolvers.Query_character(r.Context(), id)
	if err != nil {
		log.Printf("request failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("%v", friends)
	if err := json.NewEncoder(w).Encode(friends); err != nil {
		log.Printf("encoding failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetCharacters(w http.ResponseWriter, r *http.Request) {
	var input []string
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Printf("decoding json: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var characters []starwars.Character
	for _, id := range input {
		char, err := baseResolvers.Query_character(r.Context(), id)
		if err != nil {
			log.Printf("request failed: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		characters = append(characters, char)
	}
	if err := json.NewEncoder(w).Encode(characters); err != nil {
		log.Printf("encoding failed: %v", err)
	}
}
