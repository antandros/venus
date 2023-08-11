package distros

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/antandros/venus/lib"
	"github.com/antandros/venus/models"
)

type Ubuntu struct {
	Versions []models.Image
}

func (ub *Ubuntu) GetVersions() []models.Image {
	return ub.Versions
}
func (ub *Ubuntu) BaseURL() string {
	return "https://cloud-images.ubuntu.com/"
}
func (ub *Ubuntu) SetVersions(versions []models.Image) {
	ub.Versions = versions
}
func (ub *Ubuntu) Name() string {
	return "Ubuntu"
}
func (ub *Ubuntu) Fetch() {
	resp, err := http.Get("https://cloud-images.ubuntu.com/daily/streams/v1/com.ubuntu.cloud:daily:download.json")
	if err != nil {
		panic(err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var data map[string]interface{}
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		panic(err)
	}
	products := lib.TMI(data["products"])
	for _, item := range products {
		version := lib.TMI(item)
		versions := lib.TMI(version["versions"])
		var reldate time.Time
		var versionItem map[string]interface{}
		for date, version := range versions {
			t, err := time.Parse("20060102", date)
			if err != nil {
				continue
			}

			if reldate.Before(t) {
				versionItem = lib.TMI(version)
				reldate = t
			}

		}

		oelTime, err := time.Parse("2006-01-02", version["support_eol"].(string))
		if err != nil {
			continue
		}
		versionItems := lib.TMI(versionItem["items"])
		var files []models.ImageFile
		for _, fileItem := range versionItems {
			file := lib.TMI(fileItem)
			files = append(files, models.ImageFile{
				Type: file["ftype"].(string),
				SHA:  file["sha256"].(string),
				MD5:  file["md5"].(string),
				Path: file["path"].(string),
				Size: file["size"].(float64),
			})
		}
		parsedVersion := models.Image{
			Arch:            version["arch"].(string),
			Version:         version["version"].(string),
			Supported:       version["supported"].(bool),
			SupportEol:      oelTime,
			Os:              version["os"].(string),
			Release:         version["release"].(string),
			ReleaseCodename: version["release_codename"].(string),
			ReleaseTitle:    version["release_title"].(string),
			Files:           files,
			BaseUrl:         ub.BaseURL(),
		}
		ub.Versions = append(ub.Versions, parsedVersion)
	}

}
