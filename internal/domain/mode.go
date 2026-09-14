package domain

type Mode string

const (
	ModeHijack Mode = "hijack"
	ModeRoute  Mode = "route"
	ModeNormal Mode = "normal"
)

func (m Mode) Valid() bool {
	switch m {
	case ModeHijack, ModeRoute, ModeNormal:
		return true
	default:
		return false
	}
}
