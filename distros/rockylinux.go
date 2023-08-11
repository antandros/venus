package distros

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/antandros/venus/models"
)

type RockyLinux struct {
	Versions []models.Image
}

func (rck *RockyLinux) GetVersions() []models.Image {
	return rck.Versions
}

func (rck *RockyLinux) SetVersions(versions []models.Image) {
	rck.Versions = versions
}
func (rck *RockyLinux) Name() string {
	return "RockyLinux"
}
func (rck *RockyLinux) BaseURL() string {
	return "https://download.rockylinux.org/pub/rocky/"
}
func (rck *RockyLinux) FetchFolder(uri string, folder bool) []string {
	fmt.Println("Fetch folder", uri)
	resp, err := http.Get(uri)
	if err != nil {
		panic(err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	respData := string(respBody)
	hrefs := strings.Split(respData, "<a href")
	var files []string
	for _, href := range hrefs[1:] {
		urin := strings.Split(strings.Split(href, ">")[0], "\"")[1]
		uriFolder, _ := url.JoinPath(uri, urin)
		if !strings.EqualFold(urin[:1], ".") && strings.EqualFold(urin[len(urin)-1:], "/") && folder {
			files = append(files, uriFolder)
		} else if !folder && !strings.EqualFold(urin[:1], ".") {
			files = append(files, uriFolder)
		}

	}
	return files
}
func (rck *RockyLinux) FindImages(release string, releaseTitle string, version string, arch string, images []models.Image) *models.Image {
	for _, image := range images {
		if strings.EqualFold(image.Arch, arch) && strings.EqualFold(image.ReleaseTitle, releaseTitle) && strings.EqualFold(image.Release, release) && strings.EqualFold(image.Version, version) {
			return &image
		}
	}
	return nil
}
func (rck *RockyLinux) Fetch() {
	uri := "https://download.rockylinux.org/pub/rocky/"
	files := rck.FetchFolder(uri, true)
	var images []models.Image
	for _, file := range files {
		enVersion := path.Base(file)
		metaUri, _ := url.JoinPath(file, "/images/")

		_, err := http.Get(metaUri)
		if err != nil {
			continue
		}
		archFiles := rck.FetchFolder(metaUri, true)
		for _, archFile := range archFiles {
			fetchFiles := rck.FetchFolder(archFile, false)
			orgVersion := path.Base(archFile)
			for _, fetfetchFile := range fetchFiles {
				fileName := path.Base(fetfetchFile)
				imageParams := strings.Split(fileName, "-")

				if len(imageParams) < 2 {
					continue
				}
				release := imageParams[1]
				fileType := imageParams[2]
				fileTarget := ""
				releaseTitle := ""
				releaseCode := ""
				arch := ""
				imageType := ""
				isChecksum := false
				if len(imageParams) > 3 {

					fileTarget = imageParams[3]
					if len(imageParams) > 4 {
						imageType = imageParams[len(imageParams)-1]

					}
				}
				fileTargetParams := strings.Split(fileTarget, ".")
				fileParams := strings.Split(fileType, ".")
				if len(fileTargetParams) > 1 {
					releaseTitle = fileTargetParams[1]
					fileTarget = fileTargetParams[0]
					imageType = fileTargetParams[len(fileTargetParams)-1]
				}
				if len(fileParams) > 1 {
					fileType = fileParams[0]
					releaseTitle = fileParams[1]
					arch = fileParams[2]
					imageType = fileParams[3]
				}
				imageTypeParams := strings.Split(imageType, ".")
				if len(imageTypeParams) > 1 {
					releaseCode = imageTypeParams[0]
					imageType = imageTypeParams[len(imageTypeParams)-1]
					isChecksum = strings.EqualFold(imageType, "checksum")
					if isChecksum {
						imageType = imageTypeParams[len(imageTypeParams)-2]
					}
				}
				if arch == "" {
					arch = orgVersion
				}
				if releaseTitle == "" {
					releaseTitle = enVersion
				}
				if releaseCode == "" {
					releaseCode = enVersion
				}
				if !strings.EqualFold(fileType, "genericcloud") {
					continue
				}
				imageFounded := true
				imageItem := rck.FindImages(release, releaseTitle, enVersion, arch, images)
				if imageItem == nil {
					imageFounded = false
					imageItem = &models.Image{
						Arch:            arch,
						Release:         release,
						ReleaseTitle:    releaseTitle,
						Version:         enVersion,
						ReleaseCodename: releaseCode,
						Target:          fileTarget,
						BaseUrl:         rck.BaseURL(),
						Os:              "Rocky",
					}
				}
				founded := false
				selectedIndex := -1
				for index, file := range imageItem.Files {
					found := strings.EqualFold(file.Type, imageType)
					found = found && strings.EqualFold(file.Type, imageType)
					if found {
						founded = true
					}
					selectedIndex = index
				}
				itemUrlPath := strings.ReplaceAll(fetfetchFile, rck.BaseURL(), "")
				isChecksum = isChecksum || strings.EqualFold(imageType, "checksum")
				isChecksum = isChecksum || strings.EqualFold(itemUrlPath[len(itemUrlPath)-len("CHECKSUM"):], "checksum")
				if isChecksum {
					if founded {
						imageItem.Files[selectedIndex].CHECKSUM = fetfetchFile
					} else {
						imageItem.Files = append(imageItem.Files, models.ImageFile{
							CHECKSUM: fetfetchFile,
						})
					}
				} else {
					if founded {
						imageItem.Files[selectedIndex].Path = itemUrlPath
						imageItem.Files[selectedIndex].Type = imageType
					} else {
						imageItem.Files = append(imageItem.Files, models.ImageFile{
							Type: imageType,
							Path: itemUrlPath,
						})
					}
				}
				if !imageFounded {
					images = append(images, *imageItem)
				}
				rck.Versions = images

			}
		}
	}

}
