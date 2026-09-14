package domain

import (
	"fmt"
	"strings"
)

type Rule struct {
	Domain string `yaml:"domain" json:"domain"`
	Mode   Mode   `yaml:"mode" json:"mode"`
}

func (r Rule) Normalize() Rule {
	r.Domain = normalizeDomain(r.Domain)
	return r
}

func (r Rule) Validate() error {
	if r.Domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	if !r.Mode.Valid() {
		return fmt.Errorf(
			"invalid mode %q for domain %q",
			r.Mode,
			r.Domain,
		)
	}

	return nil
}

func normalizeDomain(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".")
	value = strings.ToLower(value)

	return value
}
