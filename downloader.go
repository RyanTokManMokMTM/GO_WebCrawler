package main

import (
	webC "GO_WebCrawler/webCrawler"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type downloadInfo struct {
	MovieId    uint
	YoutubeKey string
}

func readMovieJSON(path string) []*downloadInfo {
	jsonFile, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(err)
	}

	var movieInfo webC.MovieInfo
	err = json.Unmarshal(jsonFile, &movieInfo)
	if err != nil {
		log.Fatalln(err)
	}

	var movieRelatedTrailers []*downloadInfo
	//TODO - Read all the available key and append to the list
	for _, info := range movieInfo.VideoInfos.Results {
		if info.Type == "Trailer" || info.Site == "Youtube" {
			trailerInfo := downloadInfo{
				MovieId:    movieInfo.Id,
				YoutubeKey: info.Key,
			}
			movieRelatedTrailers = append(movieRelatedTrailers, &trailerInfo)
		}
	}

	return movieRelatedTrailers
}

//TODO - VideoDownloader Concurrency Starting point

func VideoDownloader(filePath string, db *gorm.DB) {
	var allVideoInfo []*downloadInfo
	fileDir, err := os.ReadDir(filePath)
	if err != nil {
		log.Fatalln(err)
	}

	for _, json := range fileDir {
		fileLoc := fmt.Sprintf("%s/%s", filePath, json.Name())
		allVideoInfo = append(allVideoInfo, readMovieJSON(fileLoc)...)
	}
	asyncDownloader(allVideoInfo, db)
}

func asyncDownloader(downloadData []*downloadInfo, db *gorm.DB) {
	downloaderCh := make(chan *downloadInfo, 20) //received downloadInfo to download
	finishedCh := make(chan *downloadInfo, 20)   //received downloadInfo after is done
	wg := sync.WaitGroup{}
	go isDone(finishedCh, db)

	go func() {
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go downloaderHandle(&wg, downloaderCh, finishedCh)
		}

	}()

	for _, info := range downloadData {
		downloaderCh <- info //each time 10 videos
	}

	defer close(finishedCh)
	close(downloaderCh)
	wg.Wait()
	log.Println("Video is finished download")
}

func downloaderHandle(wg *sync.WaitGroup, downloadCH chan *downloadInfo, isDone chan *downloadInfo) {
	defer wg.Done()
	for {
		v, ok := <-downloadCH
		if !ok {
			//fmt.Println("channel is already closed")
			break
		}
		isDone <- cmdDownloader(v)
	}
}

func isDone(ch chan *downloadInfo, db *gorm.DB) {
	for {
		v, ok := <-ch
		if !ok {
			fmt.Println("channel is already closed")
			break
		}
		if v != nil {
			fmt.Println(v.YoutubeKey)
			db.Create(webC.MovieVideoInfo{
				MovieID:  v.MovieId,
				FilePath: fmt.Sprintf("/%s.mp4", v.YoutubeKey),
			})
			os.Remove(fmt.Sprintf("D:/datas/movies/%d.json", v.MovieId))
		}
	}
}

func cmdDownloader(info *downloadInfo) *downloadInfo {
	output := fmt.Sprintf("D:/datas/trailer/%s.mp4", info.YoutubeKey)
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", info.YoutubeKey)
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("yt-dlp.exe", "-o", output, url, "--external-downloader", "aria2c", "--external-downloader-args", "-x 16 -k 1M")
	//outPipeline, err := cmd.StdoutPipe()
	//if err != nil {
	//	return nil
	//}
	//
	//errPipeline, err := cmd.StderrPipe()
	//if err != nil {
	//	return nil
	//}
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		return nil
	}
	//go cmdOutputS(outPipeline)
	//go cmdOutputS(errPipeline)
	cmd.Wait()
	errStr := stderr.String()
	errors := strings.Split(errStr, "\n")

	if len(errors) > 2 {
		fmt.Println(errStr)
		return nil
	}
	return info
}

func cmdOutputS(r io.Reader) {
	scanner := bufio.NewScanner(r) // a scanner read data form r
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
