package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func main() {
	r := chi.NewRouter()

	r.Get("/urlshortner", UrlShorter)
	r.Get("/origninalurl", GetOGURL)

	http.ListenAndServe(":3000", r)
}

func RandString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano())) // Seed the random number generator
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return fmt.Sprintf("http://%s.com", string(b))
}

var ctx = context.Background()

type requestBody struct {
	UrlLink string `json:"url_link"`
}

func UrlShorter(w http.ResponseWriter, r *http.Request) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("address"),
		Username: "default",
		Password: os.Getenv("PASSWORD"), // no password set
		DB:       0,  // use default DB
	})
	var reqBody requestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	shortURL := RandString(6)
	err = rdb.Set(ctx, shortURL, reqBody.UrlLink, 0).Err()
	if err != nil {
		panic(err)
	}
	w.Write([]byte(shortURL))

}

type shortUrlReqBody struct {
	ShortURL string `json:"short_url"`
}

func GetOGURL(w http.ResponseWriter, r *http.Request) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("address"),
		Username: "default",
		Password: os.Getenv("PASSWORD"), // no password set
		DB:       0,  // use default DB
	})
	var reqBody shortUrlReqBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	longURL, err := rdb.Get(ctx, reqBody.ShortURL).Result()
	if err == redis.Nil {
		http.Error(w, "Key doesn't exist", http.StatusBadRequest)
	} else if err != nil {
		panic(err)
	} else {
		w.Write([]byte(longURL))
	}
}
