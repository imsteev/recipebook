package controllers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/imsteev/recipebook/middleware"
	"github.com/imsteev/recipebook/models"
	"github.com/imsteev/recipebook/views"
	"gorm.io/gorm"
)

type RecipebookController struct {
	DB     *gorm.DB
	Engine *views.Engine
	Store  sessions.Store
}

func (c *RecipebookController) NewRecipeBook(w http.ResponseWriter, r *http.Request) {
	c.Engine.Render(w, "recipebooks-new.html", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	})
}

func (c *RecipebookController) CreateRecipeBook(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recipebook := models.RecipeBook{
		Name:      r.FormValue("name"),
		CreatedBy: r.Context().Value(middleware.LoggedInUserCtxKey{}).(uint),
	}
	c.DB.Create(&recipebook)
	w.Header().Add("HX-Redirect", fmt.Sprintf("/recipebooks/%d", recipebook.ID))

}

func (c *RecipebookController) ListRecipebooks(w http.ResponseWriter, r *http.Request) {
	var recipebooks []models.RecipeBook
	fmt.Println(r.Context().Value(middleware.LoggedInUserCtxKey{}))
	err := c.DB.Find(&recipebooks).Where("created_by = ?", r.Context().Value(middleware.LoggedInUserCtxKey{}).(uint)).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Engine.Render(w, "recipebooks-list.html", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
		"RecipeBooks":    recipebooks,
	})
}

func (c *RecipebookController) GetRecipeBook(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	recipeBookId := params["id"]

	var recipebook models.RecipeBook
	err := c.DB.Where("id = ?", recipeBookId).First(&recipebook).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Engine.Render(w, "recipebooks-show.html", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
		"RecipeBook":     recipebook,
	})
}
