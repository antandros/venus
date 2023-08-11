package lib

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

type Downloader struct {
	downloading chan int64
	total       int64
	received    int64
	percentage  float64
	percentFunc func(float64, int64, int64)
	doneFunc    func(string, int64, time.Duration)
}

func (d *Downloader) Wait() {

	if d != nil {
		if d.downloading != nil {
			<-d.downloading
		} else {
			fmt.Println("d.downloading is nil")
		}
	} else {
		fmt.Println("d is nil")
	}

}
func (d *Downloader) SendDownloadPercent(percent func(float64, int64, int64), path string, total int64) {
	var stop bool = false
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	for {
		select {
		case <-d.downloading:
			stop = true
		default:
			fi, err := file.Stat()
			if err != nil {
				d.downloading <- 1
				log.Fatal(err)
			}

			size := fi.Size()
			if size == 0 {
				size = 1
			}

			var percentCalculated float64 = float64(size) / float64(total) * 100
			percent(percentCalculated, size, total)
		}

		if stop {
			break
		}
		time.Sleep(time.Millisecond * 500)
	}
}

func (d *Downloader) SendPerc(perc float64, received int64, total int64) {
	d.received = received
	d.percentage = perc
	d.total = total
	if d.percentFunc != nil {

		d.percentFunc(perc, received, total)
	}
}
func (d *Downloader) SendDone(path string, total int64, duration time.Duration) {

	if d.doneFunc != nil {

		d.doneFunc(path, total, duration)
	}
}
func StartDownload(url string, dest string, percentFunc func(float64, int64, int64), doneFunc func(string, int64, time.Duration)) (*Downloader, error) {
	downloader := &Downloader{
		downloading: make(chan int64),
	}
	go func() {
		err := downloader.Download(url, dest, percentFunc, doneFunc)
		if err != nil {
			panic(err)
		}
	}()

	return downloader, nil

}
func (d *Downloader) Download(url string, dest string, percentFunc func(float64, int64, int64), doneFunc func(string, int64, time.Duration)) error {
	d.downloading = make(chan int64)
	file := path.Base(url)
	d.percentFunc = percentFunc
	d.doneFunc = doneFunc

	filePath := path.Join(dest, "/", file)
	start := time.Now()

	out, err := os.Create(filePath)

	if err != nil {
		d.downloading <- 1
		return err
	}

	defer out.Close()
	headResp, err := http.Head(url)

	if err != nil {
		d.downloading <- 1
		return err
	}

	defer headResp.Body.Close()

	size, err := strconv.Atoi(headResp.Header.Get("Content-Length"))

	if err != nil {
		return err
	}

	go d.SendDownloadPercent(d.SendPerc, filePath, int64(size))

	resp, err := http.Get(url)

	if err != nil {
		d.downloading <- 1
		panic(err)
	}

	defer resp.Body.Close()

	n, err := io.Copy(out, resp.Body)

	if err != nil {
		d.downloading <- 1
		return err
	}

	d.downloading <- n

	elapsed := time.Since(start)
	close(d.downloading)
	d.SendDone(filePath, int64(size), elapsed)
	return nil
}
