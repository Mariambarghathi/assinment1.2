package main

import (
	"backend-project/controller"
	"errors"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/go-michi/michi"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// CORS Middleware to set CORS headers
func enableCorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the necessary CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests (OPTIONS method)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// For other requests, continue to the next handler
		next.ServeHTTP(w, r)
	})
}

func main() {

	//database connection
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	controller.SetDB(db)

	//auto migration
	mig, err := migrate.New(
		"file://"+GetRootpath(os.Getenv("MIGRATIONS_ROOT")),
		os.Getenv("DATABASE_URL"),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := mig.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			log.Fatal(err)
		}
		log.Printf("migrations: %s", err.Error())
	}

	//routing
	r := michi.NewRouter()
	r.Route("/", func(sub *michi.Router) {
		//USERS
		sub.HandleFunc("GET users", controller.IndexUserHandler)
		sub.HandleFunc("GET users/{id}", controller.ShowUserHandler)
		sub.HandleFunc("POST users", controller.SignUpHandler)
		sub.HandleFunc("PUT users/{id}", controller.UpdateUserHandler)
		sub.HandleFunc("DELETE users/{id}", controller.DeleteUserHandler)
		sub.HandleFunc("POST login", controller.LoginUserHandler)

		//VENDORS
		sub.HandleFunc("GET vendor", controller.IndexVendorHandler)
		sub.HandleFunc("GET vendor/{name}", controller.ShowVendorHandler)
		sub.HandleFunc("POST vendor", controller.SaveVendorHandler)
		sub.HandleFunc("PUT vendor/{name}", controller.UpdateVendorHandler)
		sub.HandleFunc("DELETE vendor/{name}", controller.DeleteVendorHandler)
	})
	// Wrap router with the CORS middleware
	http.ListenAndServe(":8000", enableCorsMiddleware(r))
}

// get root path
func GetRootpath(dir string) string {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	return path.Join(path.Dir(ex), dir)
}
