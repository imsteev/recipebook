package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/imsteev/recipebook/models"
	"github.com/imsteev/recipebook/views"
	"gorm.io/gorm"
)

type RecipeController struct {
	DB     *gorm.DB
	Engine *views.Engine
	Store  *sessions.CookieStore
}

func (c *RecipeController) NewRecipe(w http.ResponseWriter, r *http.Request) {
	err := c.Engine.ExecuteContent(w, "recipe-form.html", map[string]string{
		"Title":  "New Recipe",
		"Action": "/recipes",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *RecipeController) CreateRecipe(w http.ResponseWriter, r *http.Request) {
	sesh, err := c.Store.Get(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sesh.Values["loggedInUserID"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	recipe := models.Recipe{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}
	if recipe.Name == "" || recipe.Description == "" {
		http.Error(w, "Name and description are required", http.StatusBadRequest)
		return
	}

	if err := c.DB.Create(&recipe).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}

func (c *RecipeController) ListRecipes(w http.ResponseWriter, r *http.Request) {
	sesh, err := c.Store.Get(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sesh.Values["loggedInUserID"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var recipes []models.Recipe
	c.DB.Find(&recipes)

	err = c.Engine.ExecuteContent(w, "recipes.html", recipes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *RecipeController) GetRecipe(w http.ResponseWriter, r *http.Request) {
	sesh, err := c.Store.Get(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sesh.Values["loggedInUserID"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	params := mux.Vars(r)
	recipeID := params["id"]

	var recipe models.Recipe
	if err := c.DB.Preload("Ingredients", func(db *gorm.DB) *gorm.DB {
		return db.Order("ingredients.name ASC")
	}).First(&recipe, recipeID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = c.Engine.ExecuteContent(w, "recipe.html", recipe)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *RecipeController) EditRecipe(w http.ResponseWriter, r *http.Request) {
	sesh, err := c.Store.Get(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sesh.Values["loggedInUserID"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	params := mux.Vars(r)
	recipeID := params["id"]

	var recipe models.Recipe
	if err := c.DB.Preload("Ingredients", func(db *gorm.DB) *gorm.DB {
		return db.Order("ingredients.name ASC")
	}).First(&recipe, recipeID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = c.Engine.ExecuteContent(w, "recipe-form.html", map[string]any{
		"Title":  "Edit Recipe",
		"Action": fmt.Sprintf("/recipes/%s/edit", recipeID),
		"Recipe": recipe,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (c *RecipeController) UpdateRecipe(w http.ResponseWriter, r *http.Request) {
	sesh, err := c.Store.Get(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sesh.Values["loggedInUserID"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)
	recipeID := params["id"]

	var recipe models.Recipe
	if err := c.DB.First(&recipe, recipeID).Error; err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
	}
	// Delete all existing ingredients and recipe_ingredients for this recipe
	if err := c.DB.Model(&recipe).Association("Ingredients").Clear(); err != nil {
		http.Error(w, "Failed to clear existing ingredients: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete all recipe_ingredients for this recipe
	if err := c.DB.Where("recipe_id = ?", recipe.ID).Delete(&models.RecipeIngredient{}).Error; err != nil {
		http.Error(w, "Failed to delete recipe_ingredients: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var (
		name         = r.PostFormValue("name")
		description  = r.PostFormValue("description")
		instructions = r.PostFormValue("instructions")
	)

	ingredients := c.ParseIngredients(r.PostForm["ingredients"], r.PostForm["quantities"])

	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	recipe.Name = name
	recipe.Description = description
	recipe.Instructions = instructions
	recipe.Ingredients = ingredients

	if err := c.DB.Save(&recipe).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("/recipes/%s", recipeID))
}

func (c *RecipeController) ParseIngredients(strIngredients []string, strQuantities []string) []models.Ingredient {
	var ingredientList []models.Ingredient

	for i := 0; i < len(strIngredients); i++ {
		ingredient := strings.TrimSpace(strIngredients[i])
		quantity := strings.TrimSpace(strQuantities[i])

		if ingredient == "" {
			// ignore empty ingredients. no quantity is fine.
			continue
		}

		ingredientList = append(ingredientList, models.Ingredient{Name: ingredient, Quantity: quantity})
	}

	return ingredientList
}

func (c *RecipeController) AddIngredientToRecipe(w http.ResponseWriter, r *http.Request) {
	sesh, err := c.Store.Get(r, "sesh")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if sesh.Values["loggedInUserID"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	params := mux.Vars(r)
	recipeID, _ := strconv.ParseUint(params["id"], 10, 32)

	var recipeIngredient models.RecipeIngredient
	if err := json.NewDecoder(r.Body).Decode(&recipeIngredient); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	recipeIngredient.RecipeID = uint(recipeID)
	if err := c.DB.Create(&recipeIngredient).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var ingredient models.Ingredient
	c.DB.First(&ingredient, recipeIngredient.IngredientID)

	ingredientHTML := fmt.Sprintf(`<input type="text" name="ingredients" placeholder="Ingredients" value="%s" />`, ingredient.Name)
	w.Write([]byte(ingredientHTML))
}

func (c *RecipeController) RemoveIngredientFromRecipe(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	recipeID := params["id"]
	ingredientID := params["ingredientId"]

	if err := c.DB.Where("recipe_id = ? AND ingredient_id = ?", recipeID, ingredientID).Delete(&models.RecipeIngredient{}).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
