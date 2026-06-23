package question

import (
	"github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"
)

// LaTeX represents metadata about a question that contains LaTeX content.
type LaTeX struct {
	Enabled bool
	// RawContent is the original LaTeX string stored in the DB.
	// Frontend must render this using KaTeX or MathJax.
	RawContent string
}

// ExtractLatex returns the LaTeX metadata for a question.
// Frontend rendering guidance:
//   - Use KaTeX for performance, MathJax for compatibility.
//   - Inline math: $...$   or \(...\)
//   - Block math:  $$...$$ or \[...\]
func ExtractLatex(q *models.Question) LaTeX {
	return LaTeX{
		Enabled:    q.LatexEnabled,
		RawContent: q.Text,
	}
}

// HasLatex returns true if any of the question's text or options contain
// common LaTeX delimiters, useful for auto-detection.
func HasLatex(text string) bool {
	for i := 0; i < len(text)-1; i++ {
		if text[i] == '$' {
			return true
		}
		if text[i] == '\\' && (text[i+1] == '(' || text[i+1] == '[') {
			return true
		}
	}
	return false
}
