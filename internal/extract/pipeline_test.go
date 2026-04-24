package extract

import (
	"testing"

	"tangled.org/dunkirk.sh/pear/internal/extract/generic"
	"tangled.org/dunkirk.sh/pear/internal/extract/schema"
)

func TestPipelineNutritionStripped(t *testing.T) {
	const html = `<!DOCTYPE html>
<html><head>
<script type="application/ld+json">
{"@context":"https://schema.org","@graph":[{"@type":"Article","headline":"Test"},{"@type":"WebPage","name":"Test"}]}
</script>
<script type="application/ld+json">
{
  "@context": "http://schema.org",
  "@type": "Recipe",
  "name": "Creamy Ginger Green Smoothie",
  "description": "A creamy smoothie",
  "recipeIngredient": ["2 handfuls organic spinach","1 cup filtered water"],
  "recipeInstructions": []
}
</script>
</head><body>
<h1 class="recipe-title">Creamy Ginger Green Smoothie</h1>
<div class="ingredients">
<p>2 handfuls organic spinach</p>
<p>1 cup filtered water</p>
</div>
<div class="directions">
<div class="instructions">
<p>Add all ingredients to a blender and blend until creamy.</p>
<p>Enjoy!</p>
</div>
</div>
</body></html>`

	recipe, ok := schema.Extract(html)
	if !ok {
		t.Fatal("JSON-LD extractor should have found the recipe")
	}
	if len(recipe.Instructions) != 0 {
		t.Errorf("expected no instructions from JSON-LD, got %d", len(recipe.Instructions))
	}
	t.Logf("JSON-LD: Name=%q Instructions=%d Ingredients=%d", recipe.Name, len(recipe.Instructions), len(recipe.Ingredients))

	recipe2, ok2 := generic.Extract(html)
	if !ok2 {
		t.Fatal("generic extractor should have found the recipe")
	}
	if len(recipe2.Instructions) == 0 {
		t.Error("generic extractor should have found instructions")
	}
	if len(recipe2.Ingredients) == 0 {
		t.Error("generic extractor should have found ingredients")
	}
	t.Logf("Generic: Name=%q Instructions=%d Ingredients=%d", recipe2.Name, len(recipe2.Instructions), len(recipe2.Ingredients))
}
