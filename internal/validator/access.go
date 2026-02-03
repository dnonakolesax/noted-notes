package validator

import "regexp"

type ACLValidator struct {
	re *regexp.Regexp
}

func NewACL() *ACLValidator {
	return &ACLValidator{re: regexp.MustCompile(`^(r)(w|\-)(x|\-)$`)}
}

func (av *ACLValidator) Validate(access string) bool {
	if len(access) != 3{
		return false
	}
	return av.re.MatchString(access)
}
