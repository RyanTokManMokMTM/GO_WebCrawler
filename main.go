package main

import (
	"WebCrawler/webCrawler"
	"fmt"
	"strconv"
)

const (
	host string = "https://api.themoviedb.org/3"
	apiKey string = "29570e7acc52b3e085ab46f6a60f0a55"
	upcomingURI = "/movie/upcoming"
)

func main(){

	var uris []string
	var res []*webCrawler.APIResult
	uri := host + upcomingURI +"?api_key=" + apiKey
	page := webCrawler.FetchPageInfo(uri)
	uris = uriGenerator(uri,page)
	res = webCrawler.FetchMovieInfos(uris)
	fmt.Println(res[0].MoviesInfo[0].Id)
}

func uriGenerator(uri string,page int) []string{
	var uris []string
	for i := 0;i<page;i++{
		newURI := uri + "&page=" + strconv.Itoa(i+1)
		uris = append(uris, newURI)
	}

	return uris
}

