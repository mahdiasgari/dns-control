package domain

import "strings"

type Rule struct {
	Domain string `yaml:"domain" json:"domain"`
	Mode   Mode   `yaml:"mode" json:"mode"`
}

func (r Rule) Normalize() Rule {
	r.Domain = normalizeDomain(r.Domain)
	return r
}

func normalizeDomain(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".")
	value = strings.ToLower(value)

	return value
}
