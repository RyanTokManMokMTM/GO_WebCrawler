package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"strconv"

	"httpGetter/GzFileDownloader"
	"httpGetter/webCrawler"
)

const (
	host string = "https://api.themoviedb.org/3"
	apiKey string = "29570e7acc52b3e085ab46f6a60f0a55"
	upcomingURI = "/movie/upcoming"
	allMovieURI string ="/discover/movie"
	moviePopular string = "/movie/popular"
	topRate string = "/movie/top_rated"
	//upComing string = "/movie/upcoming"
	genreAllURI string = "/genre/movie/list"

	peoplePopular string = "/person/popular"

	//JSON GZ
	fileHost string = "http://files.tmdb.org/p/exports"

	sqlHOST string = "127.0.0.1"
	userName string = "postgres"
	password string = "jackson"
	port int = 5432
	db string = "testDB"
)

var (
	year int = 2021
	month int = 10
	day int = 16

	movieGZ string = fmt.Sprintf("/movie_ids_%d_%d_%d.json.gz",month,day,year)
	peopleGZ string = fmt.Sprintf("/person_ids_%d_%d_%d.json.gz",month,day,year)
)

func dbConfigure() string{

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s",userName,password,sqlHOST,port,db)
	//return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d ",sqlHOST,userName,password,db,port)
}



func main(){
	config := dbConfigure()
	fmt.Println(config)
	db, err := gorm.Open(postgres.Open(config),&gorm.Config{
		//some config here....
	})
	if err != nil {
		log.Println(err)
		return
	}

	//create table

	db.AutoMigrate(&webCrawler.GenreInfo{})
	db.AutoMigrate(&webCrawler.MovieInfo{})
	//db.AutoMigrate(&webCrawler.PersonInfo{})
	//db.AutoMigrate(&webCrawler.KnowFor{})


	//
	////downloadJSONFileZip()
	////TODO - Get Genre And Movie
	//start := time.Now()
	//genreAndMoviesAll(db)
	//end := time.Now()
	//
	//fmt.Printf("Total time is used %v",end.Sub(start))
	////TODO - Get ALL person
	////peopleAll(db)

	insertJSONsToDB("G:\\moviesData",db)
}

func allMovieIds() error{
	fileURI := fileHost + movieGZ
	_, err := GzFileDownloader.DownloadGZFile(fileURI)
	if err != nil {
		log.Println(err)
		return err
	}


	return nil
}

func genreAndMoviesAll(db *gorm.DB){
	apiURL := host + genreAllURI +"?api_key=" + apiKey + "&language=zh-TW"
	//popularUri := host + moviePopular +"?api_key=" + apiKey + "&language=zh-TW"
	//topRateUri := host + topRate + "?api_key=" + apiKey + "&language=zh-TW"

	//TODO - Insert Data to Database
	_ ,err := webCrawler.GenreTableCreate(apiURL,db)
	if err != nil{
		log.Fatalln(err)
		return
	}

	fetchMovieViaID(db)
	//making a function to handle fetching movies for genre
	//genreAll(genreList,db)
	//popularAll(popularUri,db)
	//topRageAll(topRateUri,db)
}

func fetchMovieViaID(db *gorm.DB) error {
	uri := fileHost + movieGZ
	var uris []int
	moviesData, err := GzFileDownloader.DownloadGZFile(uri)
	if err != nil {
		log.Println(err)
		return err
	}

	for _,movie := range *moviesData{
		uris = append(uris,movie.MovieID)
	}

	webCrawler.FetchMovieInfosViaIDS(uris,db)

	return nil
}

func peopleAll(db *gorm.DB){
	//uri
	apiURL := host + peoplePopular + "?api_key=" + apiKey + "&language=zh-TW"
	page := webCrawler.FetchPageInfo(apiURL)
	uris := uriGenerator(apiURL,page)
	webCrawler.FetchMovieInfos(uris,db,"people")
}

func genreAll(genreList []webCrawler.GenreInfo ,db *gorm.DB){
	//for each genreList
	// https://api.themoviedb.org/3/discover/movie?api_key=29570e7acc52b3e085ab46f6a60f0a55&language=zh-TW&sort_by=popularity.desc&page=1&with_genres=28&with_watch_monetization_types=flatrate
	//fetechingURI := host + allMovieURI + "?api_key=" + apiKey + "&language=zh-TW&sort_by=popularity.desc&page=1&with_genres="+strconv.Itoa(int(genreID))+"&with_watch_monetization_types=flatrate"
	var genreALLURI []string
	for _, genre := range genreList{
		genreID := genre.Id
		moviesUri := host + allMovieURI + "?api_key=" + apiKey + "&language=zh-TW&sort_by=popularity.desc&page=1&with_genres="+strconv.Itoa(int(genreID))+"&with_watch_monetization_types=flatrate"

		currentGenrePage := webCrawler.FetchPageInfo(moviesUri)
		list := uriGenerator(moviesUri,currentGenrePage)
		genreALLURI = append(genreALLURI,list...)
	}

	webCrawler.FetchMovieInfos(genreALLURI,db,"genre")
}

func popularAll(uri string,db* gorm.DB) {
	var uris []string
	page := webCrawler.FetchPageInfo(uri)
	uris = uriGenerator(uri,page)
	webCrawler.FetchMovieInfos(uris,db,"movie")
}

func topRageAll(uri string,db *gorm.DB){
	var uris []string
	page := webCrawler.FetchPageInfo(uri)
	uris = uriGenerator(uri,page)
	webCrawler.FetchMovieInfos(uris,db,"movie")
}

func uriGenerator(uri string,page int) []string{
	var uris []string
	for i := 0;i<page;i++{
		newURI := uri + "&page=" + strconv.Itoa(i+1)
		uris = append(uris, newURI)
	}

	return uris
}

func insertJSONsToDB(dirPath string,db *gorm.DB){
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return
	}

	for _,file := range dir{
		var movieInfo webCrawler.MovieInfo
		jsonloc := fmt.Sprintf("%s/%s",dirPath,file.Name())
		jsonsData, err := ioutil.ReadFile(jsonloc)
		if err != nil {
			log.Println(err)
			return
		}

		err = json.Unmarshal(jsonsData, &movieInfo)
		if err != nil {
			log.Println(err)
			return
		}

		if err := db.Where("id = ?",movieInfo.Id).First(&webCrawler.MovieInfo{});err !=nil{
			if errors.Is(err.Error,gorm.ErrRecordNotFound){
				//not found the record
				//insert to db
				db.Create(&movieInfo)
			}else{
				fmt.Println("???")
			}
		}
	}

}