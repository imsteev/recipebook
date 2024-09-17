package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/imsteev/recipebook/models"
	"github.com/imsteev/recipebook/views"
	"gorm.io/gorm"
)

type RecipeCommentsController struct {
	DB     *gorm.DB
	Engine *views.Engine
}

func (c *RecipeCommentsController) CreateComment(w http.ResponseWriter, r *http.Request) {
	recipeID, err := strconv.ParseUint(r.FormValue("recipe_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}
	comment := models.RecipeMessage{
		From:     r.FormValue("from"),
		RecipeID: uint(recipeID),
		Message:  r.FormValue("message"),
	}

	c.DB.Create(&comment)
	http.Redirect(w, r, fmt.Sprintf("/recipes/%d", recipeID), http.StatusSeeOther)
}
