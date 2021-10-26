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

type TMDBJson struct {
	Id int `json:"id"`
}

func init(){
	//TO INIT
	client = &http.Client{}
}

func DownloadGZFile(url string) (*[]*TMDBJson,error){
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
	var data []*TMDBJson

	//first 10 record for example

	for scanner.Scan() {
		line := scanner.Text()
		jsonData, err := toJSON(line)
		if err != nil {
			log.Println(err)
			return nil,err
		}
		data = append(data,jsonData)
	}
	return &data,nil
}

func toJSON(jsonStr string) (*TMDBJson,error) {
	var data TMDBJson
	jsonBytes := []byte(jsonStr)

	//parse to json struct
	err := json.Unmarshal(jsonBytes, &data)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	return &data,nil
}
