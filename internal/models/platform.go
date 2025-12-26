package models

type Platform string

const (
	PlatformResy      Platform = "resy"
	PlatformOpenTable Platform = "opentable"
)

func (p Platform) IsValid() bool {
	switch p {
	case PlatformResy, PlatformOpenTable:
		return true
	}
	return false
}

func (p Platform) DisplayName() string {
	switch p {
	case PlatformResy:
		return "Resy"
	case PlatformOpenTable:
		return "OpenTable"
	default:
		return string(p)
	}
}
