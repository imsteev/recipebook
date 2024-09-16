package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/imsteev/recipebook-htmx/controllers"
	"github.com/imsteev/recipebook-htmx/models"
	"github.com/imsteev/recipebook-htmx/views"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TODO: don't make these globals
var db *gorm.DB
var store *sessions.CookieStore

func init() {
	// TODO: store sessions in the database?
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options = &sessions.Options{
		MaxAge:   3600,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func main() {
	// PostgreSQL connection string
	dsn := "host=localhost dbname=recipes port=5432 sslmode=disable TimeZone=UTC"

	// Use environment variable for database URL if available
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		dsn = dbURL
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Add this line to auto-migrate the User model
	db.AutoMigrate(&models.User{}, &models.Recipe{}, &models.Ingredient{}, &models.RecipeIngredient{})

	router := mux.NewRouter()

	setupRoutes(router)

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func setupRoutes(router *mux.Router) {

	engine := views.NewEngine("base.html")

	recipeController := controllers.RecipeController{DB: db, Engine: engine, Store: store}
	authController := controllers.AuthController{DB: db, Engine: engine, Store: store}

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	router.HandleFunc("/", authController.LandingPage).Methods("GET")
	router.HandleFunc("/login", authController.LoginPage).Methods("GET")
	router.HandleFunc("/login", authController.Login).Methods("POST")
	router.HandleFunc("/logout", authController.Logout).Methods("GET")
	router.HandleFunc("/signup", authController.SignupPage).Methods("GET")
	router.HandleFunc("/signup", authController.Signup).Methods("POST")
	router.HandleFunc("/recipes", recipeController.ListRecipes).Methods("GET")
	router.HandleFunc("/recipes", recipeController.CreateRecipe).Methods("POST")
	router.HandleFunc("/recipes/new", recipeController.NewRecipe).Methods("GET")
	router.HandleFunc("/recipes/{id}", recipeController.GetRecipe).Methods("GET")
	router.HandleFunc("/recipes/{id}/edit", recipeController.EditRecipe).Methods("GET")
	router.HandleFunc("/recipes/{id}/edit", recipeController.UpdateRecipe).Methods("POST")
	router.HandleFunc("/recipes/{id}/ingredients", recipeController.AddIngredientToRecipe).Methods("POST")
	router.HandleFunc("/recipes/{id}/ingredients/{ingredientId}", recipeController.RemoveIngredientFromRecipe).Methods("DELETE")
}
