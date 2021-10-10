package webCrawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"


)

var client *http.Client

func init(){
	client = &http.Client{}
}

type APIResponse struct{
	Page int `json:"page"`
	TotalPages int `json:"total_pages"`
}

type APIResult struct {
	APIResponse
	MoviesInfo []MovieInfo `json:"results"`
}

type MovieInfo struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIds         []int   `json:"genre_ids"`
	Id               int     `json:"id"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}

func FetchPageInfo(uri string) int{
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Fatalln(err)
	}

	res, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()
	body,err := ioutil.ReadAll(res.Body)
	if err != nil{
		log.Fatalln(err)
	}
	//fmt.Println(res.Header)
	var result APIResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(result)
	return result.TotalPages //get the total page of current response
}

func FetchMovieInfos(uris []string) []*APIResult{
	var results []*APIResult
	for _,uri := range uris{
		res := doRequest(uri)
		results = append(results,res)
	}
	return results
}

func doRequest(uri string) *APIResult{
	var result APIResult
	request , err := http.NewRequest("GET",uri,nil)
	if err != nil {
		log.Fatalln(err)
	}
	res, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
	}
	err = json.Unmarshal(body,&result)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(result.Page)
	return &result
}
