package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

type ShortenRequest struct {
	URL string `json:"url"`
	TTL int    `json:"ttl"` //segundos
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

func main() {
	//Conexao com o Redis
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	//Roteador HTTP
	r := chi.NewRouter()
	r.Post("/shorten", shortenHandler)
	r.Get("/{code}", redirectHandler)

	log.Println("Servidor iniciado na porta 8080")
	http.ListenAndServe(":8080", r)

}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		return
	}

	//TTL padrao 1 dia
	ttl := 24 * 3600
	if req.TTL > 0 {
		ttl = req.TTL
	}

	//incrementa o ID global do Redis
	id, err := rdb.Incr(ctx, "url:id").Result()
	if err != nil {
		http.Error(w, "Erro ao gerar o ID", http.StatusInternalServerError)
		return
	}

	code := EncodeBase62(id)
	fmt.Println("ID:", id, "-> Short code:", code)
	key := "url:" + code

	//salva o novo campo no redis com TTL
	err = rdb.Set(ctx, key, req.URL, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		http.Error(w, "Erro ao salvar no Redis", http.StatusInternalServerError)
		return
	}

	resp := ShortenResponse{ShortURL: "http://localhost:8080/" + code}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	key := "url:" + code

	url, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Erro ao acessar o Redis", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, url, http.StatusFound)

}
