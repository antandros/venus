package main

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/antandros/venus/distros"
)

func percentPrint(percent float64, received int64, total int64) {
	fmt.Println("Percent", percent, "Received", received, "/", total)
}
func donePrint(path string, total int64, duration time.Duration) {
	fmt.Println("DONE! Path", path, total, duration)
}
func errPrint(path string, err error) {
	fmt.Println("Path", path, err)
	debug.PrintStack()
}
func main() {

	//r := distros.RockyLinux{}
	//r.Fetch()
	//return
	conf := &distros.DistroConfig{
		DownloadProcessFunction: percentPrint,
		DoneFunction:            donePrint,
		ErrorFunction:           errPrint,
		MakePath:                true,
		DownloadFolder:          "./images/",
		CacheFolder:             "./temp/",
		CacheDuration:           time.Duration(24 * time.Hour),
	}
	distro := distros.NewDistros(conf)
	files := distro.Find("rocky", "x86_64", "9", "qcow2")
	for _, file := range files {
		file.Download()
		file.Wait()
	}
	//fmt.Println(distro.Find("ubuntu", "amd64", "22.04"))
	//ub := new(distros.Ubuntu)
	//fmt.Println(ub.Find("22.04", "amd64"))
}
