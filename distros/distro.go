package distros

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/antandros/venus/lib"
	"github.com/antandros/venus/models"
	"github.com/codingsince1985/checksum"
)

type DistroConfig struct {
	DownloadProcessFunction func(float64, int64, int64)
	DoneFunction            func(string, int64, time.Duration)
	ErrorFunction           func(string, error)
	DownloadFolder          string
	MakePath                bool
	CacheFolder             string
	CacheDuration           time.Duration
}

type Distros struct {
	List   map[string]models.Distro
	Config *DistroConfig
}

func NewDistros(config *DistroConfig) *Distros {
	if config == nil {
		config = new(DistroConfig)
		config.MakePath = true
		config.CacheDuration = time.Duration(time.Hour * 24)
		config.CacheFolder = "./temp/"
		config.DownloadFolder = "./distro-images/"
	}
	distroItems := map[string]models.Distro{
		"debian": &Debian{},
		"ubuntu": &Ubuntu{},
		"rocky":  &RockyLinux{},
	}
	distros := &Distros{
		List:   distroItems,
		Config: config,
	}
	for key, _ := range distroItems {

		err := distros.Load(key)
		if err != nil {
			panic(err)
		}
	}
	return distros
}

type File struct {
	models.ImageFile
	models.Image
	distro     string
	downloader *lib.Downloader
	config     *DistroConfig
}

func (fl *File) setValue(imageItem interface{}) {
	modelType := reflect.TypeOf(imageItem)
	selfValue := reflect.ValueOf(fl).Elem()
	modelValue := reflect.ValueOf(&imageItem).Elem().Elem()
	fieldLen := modelType.NumField()
	for i := 0; i < fieldLen; i++ {
		baseField := modelType.Field(i)
		field := selfValue.FieldByName(baseField.Name)
		fieldValue := modelValue.FieldByName(baseField.Name)
		if field.IsValid() {
			if fieldValue.IsValid() {

				field.Set(fieldValue)

			}
		}
	}
}
func (fl *File) GetFileChechSum(path string, sumType string) string {
	baseCsum := ""
	switch sumType {
	case "sha1":
		sha1, _ := checksum.SHA1sum(path)
		baseCsum = sha1
	case "sha256":
		sha256, _ := checksum.SHA256sum(path)
		baseCsum = sha256
	case "md5":
		md5, _ := checksum.MD5sum(path)
		baseCsum = md5
	case "crc32":
		crc32, _ := checksum.CRC32(path)
		baseCsum = crc32
	case "blake2s256":
		blake2s256, _ := checksum.Blake2s256(path)
		baseCsum = blake2s256
	}
	return baseCsum
}
func (fl *File) ControlChechsum(path string, size int64, duration time.Duration) {
	var sumType string
	var retrivedSum string
	if fl.CHECKSUM != "" {
		if strings.EqualFold(fl.CHECKSUM[:4], "http") {
			data, err := lib.GetHttpString(fl.CHECKSUM)
			if err != nil {
				if fl.config.ErrorFunction != nil {
					fl.config.ErrorFunction(path, errors.New("checksum file retrive error"))
					return
				}
				panic(err)
			}
			lineData := strings.ReplaceAll(data, "\n\n", "\n")
			lines := strings.Split(lineData, "\n")
			for _, line := range lines {
				if len(line) == 0 {
					continue
				}
				if !strings.EqualFold(line[:1], "#") {
					lineData := strings.Split(line, "=")
					params := strings.Split(lineData[0], " ")
					sumType = strings.ToLower(params[0])
					retrivedSum = strings.Trim(lineData[1], " ")

				}
			}
		} else {
			if fl.config.ErrorFunction != nil {
				fl.config.ErrorFunction(path, errors.New("unknow checksum type"))
				return
			}
			panic(errors.New("unknow checksum type"))
		}

	}

	fileSum := fl.GetFileChechSum(path, sumType)
	if retrivedSum != fileSum {
		if fl.config.ErrorFunction != nil {
			fl.config.ErrorFunction(path, errors.New("file checksum error"))
			return
		}
		panic(errors.New("file checksum error"))
	} else {
		if fl.config.DoneFunction != nil {
			fl.config.DoneFunction(path, size, duration)
		}

	}
}
func (fl *File) Download() error {

	uri, err := url.JoinPath(fl.BaseUrl, fl.Path)
	if err != nil {
		return err
	}
	fileName := path.Base(uri)

	fileFolder := path.Join(fl.config.DownloadFolder, "/", fl.distro, "/", fl.Arch, "/", fl.Version, "/")
	filePath := path.Join(fileFolder, fileName)
	stat, err := os.Stat(filePath)
	if err == nil {
		fl.ControlChechsum(filePath, stat.Size(), time.Duration(time.Second))

		return nil
	}
	_, err = os.Stat(fileFolder)
	if err != nil {
		if fl.config.MakePath {
			err = os.MkdirAll(fileFolder, 0775)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("path not found or not accesible : %s , %s", fileFolder, err.Error())
		}
	}
	go func(uri string, filePath string, percentFunc func(float64, int64, int64), doneFunc func(string, int64, time.Duration), fl *File) {
		fl.downloader, err = lib.StartDownload(uri, filePath, percentFunc, doneFunc)
		if err != nil {
			if fl.config.ErrorFunction != nil {
				fl.config.ErrorFunction(uri, err)
			} else {
				panic(err)
			}
		}
	}(uri, fileFolder, fl.config.DownloadProcessFunction, fl.ControlChechsum, fl)
	time.Sleep(time.Second * 1)
	return nil

}
func (fl *File) Wait() {
	if fl.downloader != nil {
		fl.downloader.Wait()
	}
}
func (d *Distros) Find(os string, arch string, version string, fileType string) []*File {
	itemFound := d.GetDistro(os)
	if itemFound != nil {
		item := *itemFound
		for _, verItem := range item.GetVersions() {
			if strings.EqualFold(verItem.Arch, arch) {
				ok := strings.EqualFold(verItem.Release, version) || strings.EqualFold(verItem.ReleaseCodename, version) || strings.EqualFold(verItem.ReleaseTitle, version) || strings.EqualFold(verItem.Version, version)
				if ok {
					files := verItem.Files

					var convertedFiles []*File
					for _, f := range files {

						if f.Type == fileType {
							fl := new(File)
							fl.config = d.Config
							fl.distro = item.Name()
							fl.setValue(f)
							dn := verItem
							dn.Files = dn.Files[:0]
							fl.setValue(dn)
							convertedFiles = append(convertedFiles, fl)
						}
					}
					return convertedFiles
				}
			}
		}
	}

	return nil
}
func (d *Distros) GetDistro(name string) *models.Distro {

	for key, val := range d.List {
		if strings.EqualFold(key, name) {
			return &val
		}
	}
	return nil
}
func (d *Distros) Load(name string) error {

	_, err := os.Stat(d.Config.CacheFolder)
	if err != nil {
		if d.Config.MakePath {
			os.Mkdir(d.Config.CacheFolder, 0775)
		} else {
			return err
		}
	}

	fileName := fmt.Sprintf("%s.json", name)
	filePath := path.Join(d.Config.CacheFolder, fileName)

	stat, err := os.Stat(filePath)
	if err == nil {
		createdSince := time.Since(stat.ModTime())
		clearFile := createdSince > d.Config.CacheDuration

		if clearFile {
			err = errors.New("cache expired")
		}
	}

	if err != nil {
		d.List[name].Fetch()
		versions := d.List[name].GetVersions()
		fp, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer fp.Close()

		data, err := json.Marshal(versions)
		if err != nil {
			return err
		}
		fp.Write(data)
		fp.Close()
	} else {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		var versions []models.Image
		err = json.Unmarshal(data, &versions)
		if err != nil {
			return err

		}
		d.List[name].SetVersions(versions)
	}
	d.List[name].BaseURL()
	return nil
}
