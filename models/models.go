package models

import "gorm.io/gorm"

type Recipe struct {
	gorm.Model
	UserID       uint         `json:"user_id"`
	RecipeBookID uint         `json:"recipebook_id"` // optional
	Name         string       `json:"name"`
	Ingredients  []Ingredient `json:"ingredients" gorm:"many2many:recipe_ingredients;"`
	Description  string       `json:"description"`
	Instructions string       `json:"instructions"`
}

type Ingredient struct {
	gorm.Model
	Name     string `json:"name"`
	Quantity string `json:"quantity"` // TODO: should this be well-defined or string okay?
}

type RecipeIngredient struct {
	gorm.Model
	RecipeID     uint
	IngredientID uint
	Quantity     float64 `json:"quantity"`
	Unit         string  `json:"unit"`
}

type User struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"-"` // "-" tag prevents password from being serialized to JSON
}

// RecipeBooks is a collection of recipes.
// One requirement to consider: ability to gift RecipeBooks to other users. Would need to think about ownership.
type RecipeBook struct {
	gorm.Model
	CreatedBy uint
	Name      string
}

type RecipeBookSharedLinks struct {
	gorm.Model
	RecipeBookID uint
	Slug         string `gorm:"unique"`
}

// RecipeMessage models messages associated with a RecipeBook.
// Intention is to support a gift message, but could also be used for other
// commentary on a Recipe.
type RecipeMessage struct {
	gorm.Model
	From     string // required
	RecipeID uint   //required
	Message  string // required
}
