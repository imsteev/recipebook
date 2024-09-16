package models

import "gorm.io/gorm"

type Recipe struct {
	gorm.Model
	UserID       uint         `json:"user_id"`
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
