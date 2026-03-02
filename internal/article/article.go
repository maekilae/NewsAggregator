package article

type Article struct {
	Url         string
	Path        string
	Provider    string
	Title       string `json:"title"`
	Description string
	Thumbnail   string
	Tag         string
}
