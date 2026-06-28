// Package specs looks up firearm specifications from free, key-less sources:
// Wikipedia (full-text search for the right page) and DBpedia (structured
// infobox data extracted from Wikipedia). Results are community-sourced, so
// callers should treat them as suggestions to review, not authoritative data.
package specs

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	wikiAPI       = "https://en.wikipedia.org/w/api.php"
	dbpediaSPARQL = "https://dbpedia.org/sparql"
)

// Client fetches firearm specs from Wikipedia + DBpedia.
type Client struct{ http *http.Client }

func New() *Client { return &Client{http: &http.Client{Timeout: 20 * time.Second}} }

// SearchResult is one candidate Wikipedia page.
type SearchResult struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

var tagRE = regexp.MustCompile(`<[^>]*>`)

// Search returns candidate Wikipedia pages for a firearm name.
func (c *Client) Search(ctx context.Context, q string) ([]SearchResult, error) {
	params := url.Values{
		"action":   {"query"},
		"list":     {"search"},
		"srsearch": {q},
		"srlimit":  {"8"},
		"format":   {"json"},
	}
	var payload struct {
		Query struct {
			Search []struct {
				Title   string `json:"title"`
				Snippet string `json:"snippet"`
			} `json:"search"`
		} `json:"query"`
	}
	if err := c.getJSON(ctx, wikiAPI+"?"+params.Encode(), &payload); err != nil {
		return nil, err
	}
	out := make([]SearchResult, 0, len(payload.Query.Search))
	for _, s := range payload.Query.Search {
		desc := html.UnescapeString(tagRE.ReplaceAllString(s.Snippet, ""))
		out = append(out, SearchResult{Title: s.Title, Description: strings.TrimSpace(desc)})
	}
	return out, nil
}

// Spec is one display row of the spec sheet.
type Spec struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Page is a normalized spec sheet for a firearm.
type Page struct {
	Title string            `json:"title"`
	URL   string            `json:"url"`
	Specs []Spec            `json:"specs"`
	Fill  map[string]string `json:"fill"` // editable boating-accident fields we can suggest
}

var (
	propLabels = map[string]string{
		"manufacturer": "Manufacturer", "cartridge": "Caliber", "caliber": "Caliber",
		"action": "Action", "feed": "Feed / capacity", "weight": "Weight",
		"length": "Length", "partLength": "Barrel length", "velocity": "Muzzle velocity",
		"rate": "Rate of fire", "sights": "Sights", "designer": "Designer",
		"produced": "Produced", "variants": "Variants",
	}
	propOrder = []string{"manufacturer", "cartridge", "caliber", "action", "feed", "weight", "length", "partLength", "velocity", "rate", "sights", "designer", "produced", "variants"}
)

// Page fetches and normalizes the DBpedia infobox for a Wikipedia page title.
func (c *Client) Page(ctx context.Context, title string) (*Page, error) {
	resource := strings.ReplaceAll(strings.TrimSpace(title), " ", "_")
	query := fmt.Sprintf(`SELECT ?p ?o WHERE { <http://dbpedia.org/resource/%s> ?p ?o . FILTER(STRSTARTS(STR(?p), "http://dbpedia.org/property/")) }`, resource)
	params := url.Values{"query": {query}, "format": {"application/sparql-results+json"}}

	var payload struct {
		Results struct {
			Bindings []struct {
				P struct {
					Value string `json:"value"`
				} `json:"p"`
				O struct {
					Value string `json:"value"`
				} `json:"o"`
			} `json:"bindings"`
		} `json:"results"`
	}
	if err := c.getJSON(ctx, dbpediaSPARQL+"?"+params.Encode(), &payload); err != nil {
		return nil, err
	}

	// First value wins per property.
	values := map[string]string{}
	for _, b := range payload.Results.Bindings {
		prop := b.P.Value[strings.LastIndex(b.P.Value, "/")+1:]
		if _, seen := values[prop]; seen {
			continue
		}
		values[prop] = cleanValue(b.O.Value)
	}

	page := &Page{
		Title: title,
		URL:   "https://en.wikipedia.org/wiki/" + url.PathEscape(resource),
		Specs: []Spec{}, // never nil → marshals to [] not null
		Fill:  map[string]string{},
	}
	for _, prop := range propOrder {
		if v := values[prop]; v != "" {
			page.Specs = append(page.Specs, Spec{Label: propLabels[prop], Value: v})
		}
	}
	page.Fill["model"] = title
	if v := values["manufacturer"]; v != "" {
		page.Fill["manufacturer"] = v
	}
	if v := firstNonEmpty(values["cartridge"], values["caliber"]); v != "" {
		page.Fill["caliber"] = v
	}
	return page, nil
}

func (c *Client) getJSON(ctx context.Context, u string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "boating-accident/1.0 (self-hosted firearm inventory)")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("specs: upstream status %d", resp.StatusCode)
	}
	return json.Unmarshal(body, dst)
}

// cleanValue turns a DBpedia resource URI into a readable label, or trims a literal.
func cleanValue(v string) string {
	const dbr = "http://dbpedia.org/resource/"
	if strings.HasPrefix(v, dbr) {
		s := strings.ReplaceAll(v[len(dbr):], "_", " ")
		if dec, err := url.PathUnescape(s); err == nil {
			s = dec
		}
		return s
	}
	return strings.TrimSpace(v)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
