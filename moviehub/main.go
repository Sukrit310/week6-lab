package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"moviehub/auth"
	"moviehub/handler"
	"moviehub/mw"
	"moviehub/store"
)

func main() {
	_, secretSetInEnvironment := os.LookupEnv("JWT_SECRET")
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal("โหลดไฟล์ .env ไม่สำเร็จ: ", err)
	}

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 && !secretSetInEnvironment {
		secretFile, err := os.ReadFile(".jwt_secret")
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Fatal("อ่านไฟล์ .jwt_secret ไม่สำเร็จ: ", err)
		}
		if err == nil {
			secret = strings.TrimSpace(string(secretFile))
		}
	}
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET ต้องยาวอย่างน้อย 32 ตัวอักษร (กำหนดใน environment หรือ .env หรือสร้างไฟล์ .jwt_secret)")
	}
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("ต่อฐานข้อมูลไม่ติด ", err)
	}

	tickets := &auth.Issuer{Secret: []byte(secret), TTL: time.Hour}
	h := &handler.Handler{Store: &store.Store{Pool: pool}, Tickets: tickets}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("GET /api/movies", h.SearchMovies)
    mux.HandleFunc("GET /api/movies/{id}", h.GetMovie)
	mux.HandleFunc("GET /api/movies/{id}/reviews", h.ListReviews)

	// log.Println("MovieHub API ฟังอยู่ที่ :8080")
	// log.Fatal(http.ListenAndServe(":8080", mw.Logging(mux)))

	origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	log.Println("MovieHub API ฟังอยู่ที่ :8080 ต้อนรับ", origins)
	log.Fatal(http.ListenAndServe(":8080", mw.Logging(mw.CORS(origins, mux))))
}
