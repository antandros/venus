package distros

import (
	"time"

	"github.com/antandros/venus/lib"
	"github.com/antandros/venus/models"
	"github.com/antandros/venus/models/archtypes"
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

	var data interface{}
	data, err := lib.GetHttpJson("https://cloud-images.ubuntu.com/daily/streams/v1/com.ubuntu.cloud:daily:download.json", data)

	if err != nil {
		panic(err)
	}
	dataItem := lib.TMI(data)
	products := lib.TMI(dataItem["products"])
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
		arch, err := archtypes.ConvertType(version["arch"].(string))
		if err != nil {
			panic(err)
		}
		parsedVersion := models.Image{
			Arch:            arch,
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
