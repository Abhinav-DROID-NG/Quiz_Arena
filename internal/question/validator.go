package question

import "github.com/Abhinav-DROID-NG/Quiz_Arena/internal/models"

// Validator validates incoming question data.
type Validator struct{}

// NewValidator creates a Validator.
func NewValidator() *Validator { return &Validator{} }

// ValidateOptions returns an error string if the question's options are invalid.
func (v *Validator) ValidateOptions(q *models.Question) string {
	if q.OptionA == "" || q.OptionB == "" || q.OptionC == "" || q.OptionD == "" {
		return "all four options (A, B, C, D) are required"
	}
	switch q.Answer {
	case "A", "B", "C", "D":
		// valid
	default:
		return "answer must be one of: A, B, C, D"
	}
	return ""
}

// ValidateText returns an error string if the question text is missing.
func (v *Validator) ValidateText(q *models.Question) string {
	if len(q.Text) == 0 {
		return "question text is required"
	}
	if len(q.Text) > 10000 {
		return "question text exceeds maximum length of 10000 characters"
	}
	return ""
}

// Validate performs a full validation of the question and returns any errors.
func (v *Validator) Validate(q *models.Question) []string {
	var errs []string
	if msg := v.ValidateText(q); msg != "" {
		errs = append(errs, msg)
	}
	if msg := v.ValidateOptions(q); msg != "" {
		errs = append(errs, msg)
	}
	return errs
}
