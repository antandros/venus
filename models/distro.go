package models

type Distro interface {
	Fetch()
	GetVersions() []Image
	SetVersions([]Image)
	BaseURL() string
	Name() string
}
