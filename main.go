package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/imsteev/recipebook/controllers"
	"github.com/imsteev/recipebook/middleware"
	"github.com/imsteev/recipebook/models"
	"github.com/imsteev/recipebook/views"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TODO: don't make these globals
var store *sessions.CookieStore

func init() {

}

func main() {
	var (
		dbURL  = os.Getenv("DATABASE_URL")
		secret = os.Getenv("SESSION_SECRET")
	)
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	if secret == "" {
		log.Fatal("SESSION_SECRET is not set")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	// TODO: handle migrations separately
	if err := db.AutoMigrate(
		&models.User{},
		&models.Recipe{},
		&models.Ingredient{},
		&models.RecipeIngredient{},
		&models.RecipeBook{},
		&models.RecipeBookSharedLinks{},
	); err != nil {
		log.Fatal("failed to migrate database")
	}

	// TODO: store sessions in the database?
	store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options = &sessions.Options{
		MaxAge:   3600,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}

	router := mux.NewRouter()
	privateRouter := router.NewRoute().Subrouter()
	privateRouter.Use(middleware.NoCache)
	privateRouter.Use(middleware.RequireAuth(store))

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// Controllers
	var (
		engine               = views.NewEngine("base.html")
		authController       = controllers.AuthController{DB: db, Engine: engine, Store: store}
		recipeController     = controllers.RecipeController{DB: db, Engine: engine, Store: store}
		recipebookController = controllers.RecipebookController{DB: db, Engine: engine, Store: store}
	)
	router.HandleFunc("/", authController.LandingPage).Methods("GET")
	router.HandleFunc("/login", authController.LoginPage).Methods("GET")
	router.HandleFunc("/login", authController.Login).Methods("POST")
	router.HandleFunc("/logout", authController.Logout).Methods("GET")
	router.HandleFunc("/signup", authController.SignupPage).Methods("GET")
	router.HandleFunc("/signup", authController.Signup).Methods("POST")
	privateRouter.HandleFunc("/recipes", recipeController.ListRecipes).Methods("GET")
	privateRouter.HandleFunc("/recipes", recipeController.CreateRecipe).Methods("POST")
	privateRouter.HandleFunc("/recipes/new", recipeController.NewRecipe).Methods("GET")
	privateRouter.HandleFunc("/recipes/{id}", recipeController.GetRecipe).Methods("GET")
	privateRouter.HandleFunc("/recipes/{id}/edit", recipeController.EditRecipe).Methods("GET")
	privateRouter.HandleFunc("/recipes/{id}/edit", recipeController.UpdateRecipe).Methods("POST")
	privateRouter.HandleFunc("/recipebooks/new", recipebookController.NewRecipeBook).Methods("GET")
	privateRouter.HandleFunc("/recipebooks", recipebookController.CreateRecipeBook).Methods("POST")
	privateRouter.HandleFunc("/recipebooks", recipebookController.ListRecipebooks).Methods("GET")
	privateRouter.HandleFunc("/recipebooks/{id}", recipebookController.GetRecipeBook).Methods("GET")
	privateRouter.HandleFunc("/recipebooks/{id}/share", recipebookController.CreateRecipeBookSharedLink).Methods("POST")

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", csrf.Protect([]byte(secret))(router)))
}
