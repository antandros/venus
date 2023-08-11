package models

import "time"

type ImageFile struct {
	Type     string  `json:"type,omitempty"`
	SHA      string  `json:"sha,omitempty"`
	MD5      string  `json:"md5,omitempty"`
	CHECKSUM string  `json:"checksum,omitempty"`
	Path     string  `json:"path,omitempty"`
	Size     float64 `json:"size,omitempty"`
}
type Image struct {
	Arch            string      `json:"arch,omitempty"`
	Os              string      `json:"os,omitempty"`
	Release         string      `json:"release,omitempty"`
	Target          string      `json:"target,omitempty"`
	ReleaseCodename string      `json:"release_codename,omitempty"`
	ReleaseTitle    string      `json:"release_title,omitempty"`
	SupportEol      time.Time   `json:"support_eol,omitempty"`
	Supported       bool        `json:"supported,omitempty"`
	Version         string      `json:"version,omitempty"`
	BaseUrl         string      `json:"baseurl,omitempty"`
	Files           []ImageFile `json:"files,omitempty"`
}
