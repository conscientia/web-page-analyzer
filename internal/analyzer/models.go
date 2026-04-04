package analyzer

// PageAnalysis holds the complete analysis of a webpage.
type PageAnalysis struct {
	URL          string   `json:"url"`
	HTMLVersion  string   `json:"html_version"`
	Title        string   `json:"title"`
	Headings     Headings `json:"headings"`
	Links        Links    `json:"links"`
	HasLoginForm bool     `json:"has_login_form"`
}

// Headings holds the heading level counters found on the page.
type Headings struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
}

// Total returns the sum of all heading levels.
func (h Headings) Total() int {
	return h.H1 + h.H2 + h.H3 + h.H4 + h.H5 + h.H6
}

// Links holds link counts and accessibility counters.
type Links struct {
	Internal     int `json:"internal"`
	External     int `json:"external"`
	Inaccessible int `json:"inaccessible"`
	Unchecked    int `json:"unchecked"` // links that could not be checked due to LINK_CHECK_TIMEOUT
}

// Total returns the total number of links found on the page.
func (l Links) Total() int {
	return l.Internal + l.External
}

// Link represents a single parsed link from the page.
type Link struct {
	URL        string
	IsInternal bool
}

// AnalysisError is returned to the client when analysis cannot proceed.
type AnalysisError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}
