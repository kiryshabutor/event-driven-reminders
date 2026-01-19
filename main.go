package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kiribu/jwt-practice/handlers"
	"github.com/kiribu/jwt-practice/middleware"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/register", handlers.Register).Methods("POST")
	r.HandleFunc("/login", handlers.Login).Methods("POST")
	r.HandleFunc("/refresh", handlers.Refresh).Methods("POST")

	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.JWTAuth)
	protected.HandleFunc("/protected", handlers.Protected).Methods("GET")
	protected.HandleFunc("/profile", handlers.Profile).Methods("GET")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	port := ":8080"
	fmt.Printf("Сервер запущен на http://localhost%s\n", port)
	fmt.Println("Доступные endpoints:")
	fmt.Println("   POST   /register  - Регистрация нового пользователя")
	fmt.Println("   POST   /login     - Логин (получение токенов)")
	fmt.Println("   POST   /refresh   - Обновление access token")
	fmt.Println("   GET    /protected - Защищенный endpoint")
	fmt.Println("   GET    /profile   - Профиль пользователя")
	fmt.Println("   GET    /health    - Health check")

	log.Fatal(http.ListenAndServe(port, r))
}
