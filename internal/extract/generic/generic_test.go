package generic

import (
	"testing"
)

const testHTML = `
<html><body>
<img itemprop="image" src="https://example.com/photo.jpg" alt="Smoothie"/>
<main class="recipe-content">
<h1 class="recipe-title">Creamy Ginger Green Smoothie</h1>
<h4 class="serves">Serves 1</h4>
<div class="ingredients">
<h3>Ingredients:</h3>
<p>2 handfuls organic <a href="/spinach/">spinach</a></p>
<p>1 cup filtered water</p>
<p>1/2 avocado</p>
<p>1 medium banana</p>
</div>
<div class="directions">
<h3>Directions:</h3>
<div class="instructions">
<p>Simply add all ingredients in a high-speed blender and blend until thick and creamy.</p>
<p>You may add ice if you'd like to chill further.</p>
<p>Enjoy!</p>
</div>
</div>
</main>
</body></html>
`

func TestNutritionStripped(t *testing.T) {
	recipe, ok := Extract(testHTML)
	if !ok {
		t.Fatal("Extract returned false")
	}
	t.Logf("Name: %q", recipe.Name)
	t.Logf("Yield: %q", recipe.Yield)
	t.Logf("ImageURL: %q", recipe.ImageURL)
	t.Logf("Ingredients (%d):", len(recipe.Ingredients))
	for i, ing := range recipe.Ingredients {
		t.Logf("  [%d] RawText=%q Name=%q", i, ing.RawText, ing.Name)
	}
	t.Logf("Instructions (%d):", len(recipe.Instructions))
	for i, instr := range recipe.Instructions {
		t.Logf("  [%d] Text=%q", i, instr.Text)
	}

	if len(recipe.Ingredients) == 0 {
		t.Error("expected ingredients, got none")
	}
	if len(recipe.Instructions) == 0 {
		t.Error("expected instructions, got none")
	}
	if recipe.Name != "Creamy Ginger Green Smoothie" {
		t.Errorf("unexpected name: %q", recipe.Name)
	}
}
