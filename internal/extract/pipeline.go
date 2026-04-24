package extract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"tangled.org/dunkirk.sh/pear/internal/extract/generic"
	"tangled.org/dunkirk.sh/pear/internal/extract/hrecipe"
	"tangled.org/dunkirk.sh/pear/internal/extract/marmiton"
	"tangled.org/dunkirk.sh/pear/internal/extract/schema"
	"tangled.org/dunkirk.sh/pear/internal/extract/wprm"
	"tangled.org/dunkirk.sh/pear/internal/models"
)

var htmlLangRe = regexp.MustCompile(`(?i)<html[^>]*\blang=["']([a-zA-Z-]+)`)

type Pipeline struct {
	client         *http.Client
	flareSolverURL string
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		flareSolverURL: func() string {
		v := os.Getenv("FLARESOLVERR_URL")
		if v == "0" || v == "" {
			return ""
		}
		if v == "1" {
			return "http://localhost:8191/v1"
		}
		return v
	}(),
	}
}

type Result struct {
	Recipe *models.Recipe
	Error  error
}

func (p *Pipeline) Extract(targetURL string) *Result {
	body, err := p.fetch(targetURL)
	if err != nil {
		if p.flareSolverURL != "" {
			flareBody, flareErr := p.fetchViaFlareSolver(targetURL)
			if flareErr != nil {
				return &Result{Error: fmt.Errorf("fetching page: %w (flaresolverr: %v)", err, flareErr)}
			}
			body = flareBody
		} else {
			return &Result{Error: fmt.Errorf("fetching page: %w", err)}
		}
	}

	lang := detectLanguage(body)

	type candidate struct {
		recipe *models.Recipe
		method string
	}
	var fallbacks []candidate

	tryExtract := func(r *models.Recipe, ok bool, method string) *Result {
		if !ok || r == nil {
			return nil
		}
		r.SourceURL = targetURL
		r.SourceDomain = domainOf(targetURL)
		r.Language = lang
		r.Normalize()
		if len(r.Instructions) > 0 {
			return &Result{Recipe: r}
		}
		fallbacks = append(fallbacks, candidate{r, method})
		return nil
	}

	if r, ok := marmiton.Extract(body); true {
		if result := tryExtract(r, ok, "marmiton"); result != nil {
			return result
		}
	}
	if r, ok := wprm.Extract(body); true {
		if result := tryExtract(r, ok, "wprm"); result != nil {
			return result
		}
	}
	if r, ok := schema.Extract(body); true {
		if result := tryExtract(r, ok, "schema.org"); result != nil {
			return result
		}
	}
	if r, ok := schema.ExtractMicrodata(body); true {
		if result := tryExtract(r, ok, "microdata"); result != nil {
			return result
		}
	}
	if r, ok := hrecipe.Extract(body); true {
		if result := tryExtract(r, ok, "h-recipe"); result != nil {
			return result
		}
	}
	if r, ok := generic.Extract(body); true {
		if result := tryExtract(r, ok, "generic"); result != nil {
			return result
		}
	}

	if len(fallbacks) > 0 {
		return &Result{Recipe: fallbacks[0].recipe}
	}

	return &Result{Error: fmt.Errorf("no recipe found on page - tried JSON-LD, microdata, and h-recipe extraction")}
}

func (p *Pipeline) fetch(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Pare/1.0 (recipe extractor; like a read-it-later service)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (p *Pipeline) fetchViaFlareSolver(targetURL string) (string, error) {
	reqBody, _ := json.Marshal(map[string]any{
		"cmd":        "request.get",
		"url":        targetURL,
		"maxTimeout": 60000,
	})

	req, err := http.NewRequest("POST", p.flareSolverURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("creating flaresolverr request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("flaresolverr request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading flaresolverr response: %w", err)
	}

	var result struct {
		Status   string `json:"status"`
		Message  string `json:"message"`
		Solution struct {
			Response string `json:"response"`
			Status   int    `json:"status"`
		} `json:"solution"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parsing flaresolverr response: %w", err)
	}

	if result.Status != "ok" {
		return "", fmt.Errorf("flaresolverr: %s", result.Message)
	}

	return result.Solution.Response, nil
}

func detectLanguage(body string) string {
	if m := htmlLangRe.FindStringSubmatch(body); len(m) >= 2 {
		return strings.ToLower(m[1])
	}
	return ""
}

func domainOf(url string) string {
	parts := strings.SplitAfter(url, "://")
	if len(parts) < 2 {
		return url
	}
	host := strings.Split(parts[1], "/")[0]
	return host
}
