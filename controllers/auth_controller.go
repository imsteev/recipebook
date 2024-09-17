package controllers

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/imsteev/recipebook/models"
	"github.com/imsteev/recipebook/views"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	DB     *gorm.DB
	Engine *views.Engine
	Store  sessions.Store
}

func NewAuthController(db *gorm.DB, engine *views.Engine, store sessions.Store) *AuthController {
	return &AuthController{DB: db, Engine: engine, Store: store}
}

func (c *AuthController) LandingPage(w http.ResponseWriter, r *http.Request) {
	sesh, err := c.Store.Get(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sesh.Values["loggedInUserID"] != nil {
		http.Redirect(w, r, "/recipes", http.StatusSeeOther)
		return
	}

	err = c.Engine.Render(w, "landing.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *AuthController) LoginPage(w http.ResponseWriter, r *http.Request) {
	sesh, err := c.Store.Get(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sesh.Values["loggedInUserID"] != nil {
		http.Redirect(w, r, "/recipes", http.StatusSeeOther)
		return
	}
	err = c.Engine.Render(w, "login.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	var user models.User
	c.DB.Where("username = ?", username).First(&user)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	sesh, err := c.Store.New(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sesh.Values["loggedInUserID"] = user.ID
	sesh.Save(r, w)

	w.Header().Add("HX-Redirect", "/recipes")
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	sesh, err := c.Store.Get(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sesh.Values["loggedInUserID"] = nil
	sesh.Options.MaxAge = -1
	sesh.Save(r, w)

	w.Header().Add("HX-Redirect", "/login")
}

func (c *AuthController) SignupPage(w http.ResponseWriter, r *http.Request) {
	err := c.Engine.Render(w, "signup.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *AuthController) Signup(w http.ResponseWriter, r *http.Request) {
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
	c.DB.Create(&user)

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
