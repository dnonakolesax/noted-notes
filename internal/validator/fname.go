package validator

import "regexp"

type FNameValidator struct {
	re *regexp.Regexp
}

func NewFName() *FNameValidator {
	return &FNameValidator{re: regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁ0-9_-]+\.goi$`)}
}

func (fv *FNameValidator) Validate(fname string) bool {
	if len(fname) < 3 || len(fname) > 32 {
		return false
	}
	return fv.re.MatchString(fname)
}
