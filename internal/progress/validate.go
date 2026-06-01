package progress

import "github.com/TrebuchetDynamics/goscrapling/internal/progress/validation"

func Validate(p *Progress) error {
	return validation.Validate(p)
}
