package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/imsteev/recipebook-htmx/controllers"
	"github.com/imsteev/recipebook-htmx/lib"
	"github.com/imsteev/recipebook-htmx/models"
	"golang.org/x/crypto/bcrypt"
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

//go:embed views/*.html
var views embed.FS

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
	recipeController := controllers.Controller{DB: db, Engine: lib.NewEngine(views, "views", "base.html"), Store: store}

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	router.HandleFunc("/", landingPage).Methods("GET")
	router.HandleFunc("/login", loginPage).Methods("GET")
	router.HandleFunc("/login", loginUser).Methods("POST")
	router.HandleFunc("/signup", signupPage).Methods("GET")
	router.HandleFunc("/signup", signupUser).Methods("POST")
	router.HandleFunc("/recipes", recipeController.ListRecipes).Methods("GET")
	router.HandleFunc("/recipes", recipeController.CreateRecipe).Methods("POST")
	router.HandleFunc("/recipes/new", recipeController.NewRecipe).Methods("GET")
	router.HandleFunc("/recipes/{id}", recipeController.GetRecipe).Methods("GET")
	router.HandleFunc("/recipes/{id}/edit", recipeController.EditRecipe).Methods("GET")
	router.HandleFunc("/recipes/{id}/edit", recipeController.UpdateRecipe).Methods("POST")
	router.HandleFunc("/recipes/{id}/ingredients", recipeController.AddIngredientToRecipe).Methods("POST")
	router.HandleFunc("/recipes/{id}/ingredients/{ingredientId}", recipeController.RemoveIngredientFromRecipe).Methods("DELETE")
}

func landingPage(w http.ResponseWriter, r *http.Request) {
	sesh, err := store.Get(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sesh.Values["loggedInUserID"] != nil {
		http.Redirect(w, r, "/recipes", http.StatusSeeOther)
	}

	err = executeContent(w, "landing.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func executeContent(w http.ResponseWriter, templateName string, data any) error {
	// TODO: parse content templates once and cache them. no need to parse on every request
	// TODO: what if we want to add nested templates?
	t, err := template.ParseFS(views, "views/base.html", "views/"+templateName)
	if err != nil {
		return err
	}
	return t.ExecuteTemplate(w, "base", data)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	sesh, err := store.Get(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sesh.Values["loggedInUserID"] != nil {
		http.Redirect(w, r, "/recipes", http.StatusSeeOther)
	}
	err = executeContent(w, "login.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	var user models.User
	db.Where("username = ?", username).First(&user)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	sesh, err := store.New(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sesh.Values["loggedInUserID"] = user.ID
	sesh.Save(r, w)

	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}
func signupPage(w http.ResponseWriter, _ *http.Request) {
	err := executeContent(w, "signup.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func signupUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	password2 := r.FormValue("password2")

	if password != password2 {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := models.User{Username: username, Password: string(passwordHash)}
	db.Create(&user)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
