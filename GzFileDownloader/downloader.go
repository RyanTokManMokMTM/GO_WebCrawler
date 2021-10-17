package GzFileDownloader

import (
	"bufio"
	"compress/gzip"
	"encoding/json"

	"log"
	"net/http"
)

var (
	client *http.Client
)

type Movie struct {
	MovieID int `json:"id"`
	Title string `json:"original_title"`
}

func init(){
	//TO INIT
	client = &http.Client{}
}

func DownloadGZFile(url string) (*[]*Movie,error){
	res, err := client.Get(url)
	if err != nil {
		log.Println(res)
		return nil ,err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("Could not pull data from TMDB ,status code %d",res.StatusCode)
		return nil,http.ErrBodyNotAllowed
	}

	reader, err := gzip.NewReader(res.Body)
	if err != nil {
		log.Println(err)
		return nil,err
	}

	//add to scanner
	scanner := bufio.NewScanner(reader)
	var movies []*Movie

	//first 10 record for example
	//var i = 0
	for scanner.Scan() {
		line := scanner.Text()
		movie, err := toJSON(line)
		if err != nil {
			log.Println(err)
			return nil,err
		}
		movies = append(movies,movie)
	}
	return &movies,nil
}

func toJSON(jsonStr string) (*Movie,error) {
	var movieData Movie
	jsonBytes := []byte(jsonStr)

	//parse to json struct
	err := json.Unmarshal(jsonBytes, &movieData)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	//formatJSON, err := json.MarshalIndent(movieData, "", "\t")
	//if err != nil {
	//	log.Println(err)
	//	return nil,err
	//}
	//fmt.Println(string(formatJSON))

	return &movieData,nil
}