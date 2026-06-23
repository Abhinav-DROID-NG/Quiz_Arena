package question

// DiagramMeta holds metadata about a diagram associated with a question.
type DiagramMeta struct {
	// URL is the publicly accessible URL to the diagram image.
	// This is served from /assets/diagrams/ locally, or a CDN in production.
	URL string
	// AltText is an accessibility description of the diagram.
	AltText string
}

// DiagramFromURL creates DiagramMeta from a stored URL.
func DiagramFromURL(url string) *DiagramMeta {
	if url == "" {
		return nil
	}
	return &DiagramMeta{URL: url}
}

// StoragePath derives a filesystem-relative storage path for a diagram filename.
// e.g. StoragePath("abc.png") → "diagrams/abc.png"
func StoragePath(filename string) string {
	return "diagrams/" + filename
}
