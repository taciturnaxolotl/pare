package models

import (
	"html"
	"time"
)

type Recipe struct {
	Name         string
	Description  string
	ImageURL     string
	SourceURL    string
	SourceDomain string
	PrepTime     string
	CookTime     string
	TotalTime    string
	Yield        string
	Servings     int
	Ingredients  []Ingredient
	Instructions []Instruction
	Language         string
	ExtractionMethod string
}

type Ingredient struct {
	RawText string
	Quantity string
	Unit     string
	Name     string
	Group    string
}

type Instruction struct {
	Text string
}

type CachedRecipe struct {
	URL        string
	Recipe     []byte
	ExtractionMethod string
	FetchedAt  time.Time
}

func (r *Recipe) Normalize() {
	r.Name = html.UnescapeString(r.Name)
	r.Description = html.UnescapeString(r.Description)
	for i := range r.Ingredients {
		r.Ingredients[i].RawText = html.UnescapeString(r.Ingredients[i].RawText)
		r.Ingredients[i].Name = html.UnescapeString(r.Ingredients[i].Name)
		r.Ingredients[i].Group = html.UnescapeString(r.Ingredients[i].Group)
	}
	for i := range r.Instructions {
		r.Instructions[i].Text = html.UnescapeString(r.Instructions[i].Text)
	}
}