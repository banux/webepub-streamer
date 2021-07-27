package searcher

type SearchResults struct {
	Query        string         `json:"query"`
	TotalResults int            `json:"totalResults"`
	Results      []SearchResult `json:"results"`
}

type SearchResult struct {
	Resource   string  `json:"resource"`
	Title      string  `json:"title"`
	Match      string  `json:"match"`
	TextBefore string  `json:"text-before,omitempty"`
	TextAfter  string  `json:"text-after,omitempty"`
	Locators   Locator `json:"locators"`
}

// TODO remove
type Locator struct {
	Cfi      string  `json:"cfi,omitempty"`
	Xpath    string  `json:"xpath,omitempty"`
	Page     int     `json:"page,omitempty"`
	Position float64 `json:"position,omitempty"`
}
