package controllers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
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
	var sharedLink models.RecipeBookSharedLink
	err = c.DB.Where("recipe_book_id = ?", recipebook.ID).First(&sharedLink).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.Engine.Render(w, "recipebooks-show.html", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
		"RecipeBook":     recipebook,
		"SharedLink":     sharedLink,
	})
}

func (c *RecipebookController) CreateRecipeBookSharedLink(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	recipeBookId := params["id"]

	var recipebook models.RecipeBook
	err := c.DB.Where("id = ?", recipeBookId).First(&recipebook).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	key := securecookie.GenerateRandomKey(32)
	sharedLink := models.RecipeBookSharedLink{
		RecipeBookID: recipebook.ID,
		Slug:         fmt.Sprintf("%x", key),
	}
	err = c.DB.Create(&sharedLink).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(sharedLink)
	w.Write([]byte(fmt.Sprintf(`<a href="/recipebooks/%s">%s</a>`, sharedLink.Slug, "Public Link")))
}

func (c *RecipebookController) GetRecipeBookBySlug(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	slug := params["slug"]

	var sharedLink models.RecipeBookSharedLink
	if err := c.DB.Where("slug = ?", slug).First(&sharedLink).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var recipebook models.RecipeBook
	if err := c.DB.Where("id = ?", sharedLink.RecipeBookID).First(&recipebook).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err := c.Engine.Render(w, "recipebooks-guest.html", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
		"RecipeBook":     recipebook,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
