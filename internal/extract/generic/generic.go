package generic

import (
	"strings"

	"tangled.org/dunkirk.sh/pear/internal/extract/schema"
	"tangled.org/dunkirk.sh/pear/internal/models"

	"golang.org/x/net/html"
)

func Extract(body string) (*models.Recipe, bool) {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, false
	}

	recipe := &models.Recipe{}
	recipe.ExtractionMethod = "generic"

	found := false

	if name := findByClass(doc, "recipe-title", "recipe-name"); name != "" {
		recipe.Name = name
		found = true
	}
	if desc := findByMetaContent(doc, "description"); desc != "" {
		recipe.Description = desc
	}
	if img := findByItempropImage(doc); img != "" {
		recipe.ImageURL = img
	} else if img := findByMetaContent(doc, "og:image"); img != "" {
		recipe.ImageURL = img
	}
	if yield := findByClass(doc, "serves"); yield != "" {
		recipe.Yield = strings.TrimPrefix(yield, "Serves ")
	}

	ingredients := collectIngredients(doc)
	if len(ingredients) > 0 {
		found = true
		for _, ing := range ingredients {
			recipe.Ingredients = append(recipe.Ingredients, schema.ParseIngredient(ing))
		}
	}

	instructions := collectInstructions(doc)
	if len(instructions) > 0 {
		found = true
		for _, instr := range instructions {
			recipe.Instructions = append(recipe.Instructions, models.Instruction{Text: instr})
		}
	}

	if !found {
		return nil, false
	}

	if recipe.Name == "" {
		if title := findByMetaContent(doc, "og:title"); title != "" {
			recipe.Name = title
		}
	}

	if recipe.Name == "" {
		return nil, false
	}

	return recipe, true
}

func findByClass(n *html.Node, classes ...string) string {
	var result string
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, class := range classes {
				if hasClass(n, class) {
					result = textContent(n)
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
			if result != "" {
				return
			}
		}
	}
	f(n)
	return result
}

func hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			for _, c := range strings.Fields(attr.Val) {
				if c == class {
					return true
				}
			}
		}
	}
	return false
}

func findByMetaContent(n *html.Node, name string) string {
	var f func(*html.Node) string
	f = func(n *html.Node) string {
		if n.Type == html.ElementNode && n.Data == "meta" {
			metaName, metaProp, metaContent := "", "", ""
			for _, attr := range n.Attr {
				switch attr.Key {
				case "name":
					metaName = attr.Val
				case "property":
					metaProp = attr.Val
				case "content":
					metaContent = attr.Val
				}
			}
			if (metaName == name || metaProp == name) && metaContent != "" {
				return metaContent
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if v := f(c); v != "" {
				return v
			}
		}
		return ""
	}
	return f(n)
}

func findByItempropImage(n *html.Node) string {
	var f func(*html.Node) string
	f = func(n *html.Node) string {
		if n.Type == html.ElementNode {
			hasImageProp := false
			for _, attr := range n.Attr {
				if attr.Key == "itemprop" && attr.Val == "image" {
					hasImageProp = true
					break
				}
			}
			if hasImageProp && n.Data == "img" {
				for _, attr := range n.Attr {
					if attr.Key == "src" {
						return attr.Val
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if v := f(c); v != "" {
				return v
			}
		}
		return ""
	}
	return f(n)
}

func collectIngredients(n *html.Node) []string {
	container := findNodeByClass(n, "ingredients", "recipe-ingredients")
	if container == nil {
		return nil
	}
	var items []string
	collectParagraphsAndListItems(container, &items)
	return items
}

func collectInstructions(n *html.Node) []string {
	container := findNodeByClass(n, "directions", "instructions", "recipe-instructions", "recipe-directions")
	if container == nil {
		return nil
	}
	var items []string
	collectParagraphsAndListItems(container, &items)
	return items
}

func findNodeByClass(n *html.Node, classes ...string) *html.Node {
	var f func(*html.Node) *html.Node
	f = func(n *html.Node) *html.Node {
		if n.Type == html.ElementNode {
			for _, class := range classes {
				if hasClass(n, class) {
					return n
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if found := f(c); found != nil {
				return found
			}
		}
		return nil
	}
	return f(n)
}

func collectParagraphsAndListItems(n *html.Node, items *[]string) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "p" || c.Data == "li") {
			text := strings.TrimSpace(textContent(c))
			if text != "" {
				*items = append(*items, text)
			}
		} else {
			collectParagraphsAndListItems(c, items)
		}
	}
}

func textContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(textContent(c))
	}
	return strings.TrimSpace(sb.String())
}
