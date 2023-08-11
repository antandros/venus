package distros

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/antandros/venus/lib"
	"github.com/antandros/venus/models"
	"github.com/jlaffaye/ftp"
)

type Debian struct {
	Versions []models.Image
}

func (deb *Debian) FetchJsonFile(path string, rawUrl string) map[string]interface{} {
	parser, _ := url.Parse(rawUrl)
	fp, err := ftp.Dial(fmt.Sprintf("%s:21", parser.Hostname()), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return deb.FetchJsonFile(path, rawUrl)
	}

	err = fp.Login("anonymous", "anonymous")
	if err != nil {
		return deb.FetchJsonFile(path, rawUrl)
	}
	r, err := fp.Retr(path)
	if err != nil {
		panic(err)
	}
	ndata, _ := io.ReadAll(r)

	var data map[string]interface{}
	err = json.Unmarshal(ndata, &data)
	if err != nil {
		panic(err)
	}
	return data
}
func (deb *Debian) FetchSub(path string, rawUrl string) []string {

	parser, _ := url.Parse(rawUrl)
	fp, err := ftp.Dial(fmt.Sprintf("%s:21", parser.Hostname()), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return deb.FetchSub(path, rawUrl)
	}

	err = fp.Login("anonymous", "anonymous")
	if err != nil {
		return deb.FetchSub(path, rawUrl)
	}
	fnx, err := fp.List(path)
	var items []string
	if err != nil {
		return deb.FetchSub(path, rawUrl)
	}
	fp.Quit()
	var folders []string
	hasLatest := false
	for _, fn := range fnx {
		if !strings.EqualFold(fn.Name, "OpenStack") && !strings.EqualFold(fn.Name, "daily") {

			if fn.Type == ftp.EntryTypeLink && strings.EqualFold(fn.Name, "latest") {
				hasLatest = true
			} else if fn.Type == ftp.EntryTypeFolder {
				folders = append(folders, fn.Name)
			} else {
				if strings.EqualFold(fn.Name[len(fn.Name)-4:], "json") && strings.Contains(fn.Name, "genericcloud") {

					pth := fmt.Sprintf("%s/%s", path, fn.Name)
					items = append(items, pth)
				}
			}
		}

	}
	if hasLatest {
		pth := fmt.Sprintf("%s/latest", path)
		return deb.FetchSub(pth, rawUrl)
	} else {
		for _, fn := range folders {
			pth := fmt.Sprintf("%s/%s", path, fn)
			items = append(items, deb.FetchSub(pth, rawUrl)...)
		}
	}

	return items
}
func (deb *Debian) GetVersions() []models.Image {
	return deb.Versions
}

func (deb *Debian) SetVersions(versions []models.Image) {
	deb.Versions = versions
}
func (deb *Debian) Name() string {
	return "Debian"
}
func (deb *Debian) BaseURL() string {
	return "https://laotzu.ftp.acc.umu.se/images/cloud/"
}
func (deb *Debian) Fetch() {
	rawURL := "ftp://cloud.debian.org/images/cloud"
	parser, _ := url.Parse(rawURL)

	pth := fmt.Sprintf(".%s", parser.Path)
	files := deb.FetchSub(pth, rawURL)
	for _, n := range files {
		resp := deb.FetchJsonFile(n, rawURL)
		items := resp["items"].([]interface{})
		var ver models.Image
		var images []models.ImageFile
		for _, item := range items {
			bitem := lib.TMI(item)
			if strings.EqualFold(bitem["kind"].(string), "upload") {
				data := lib.TMI(lib.TMI(bitem["data"]))
				metadata := lib.TMI(lib.TMI(bitem["metadata"]))
				labels := lib.TMI(lib.TMI(metadata["labels"]))
				annon := lib.TMI(lib.TMI(metadata["annotations"]))

				images = append(images, models.ImageFile{
					Type: labels["upload.cloud.debian.org/image-format"].(string),
					SHA:  annon["cloud.debian.org/digest"].(string),
					Path: data["ref"].(string),
				})
			}
			if strings.EqualFold(bitem["kind"].(string), "build") {
				info := lib.TMI(lib.TMI(bitem["data"])["info"])

				ver = models.Image{
					Arch:            info["arch"].(string),
					Os:              "debian",
					Release:         info["release"].(string),
					ReleaseCodename: info["release_id"].(string),
					ReleaseTitle:    info["release_baseid"].(string),
					Version:         info["version"].(string),
					BaseUrl:         deb.BaseURL(),
				}
			}

		}
		ver.Files = images
		deb.Versions = append(deb.Versions, ver)

	}
	// Do something with the FTP conn

}
