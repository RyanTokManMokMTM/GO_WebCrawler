package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"httpGetter/GzFileDownloader"
	"httpGetter/webCrawler"

	"io/ioutil"
	"log"
	"strconv"
	"time"
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
	year ,month,day = time.Now().Date()
	movieGZ string = fmt.Sprintf("/movie_ids_%d_%d_%d.json.gz",int(month),day,year)
	peopleGZ string = fmt.Sprintf("/person_ids_%d_%d_%d.json.gz",int(month),day,year)
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
	db.AutoMigrate(&webCrawler.PersonInfo{})
	db.AutoMigrate(&webCrawler.MovieCharacter{})
	db.AutoMigrate(&webCrawler.PersonCrew{})
	db.AutoMigrate(&webCrawler.Department{})

	//TODO - GET DEPARTMENT
	//departmentURI := fmt.Sprintf("%s/configuration/jobs?api_key=%s",host,apiKey)
	//webCrawler.FetchAllDepartment(departmentURI,db)

	//TODO - Get Genre And Movie
	//movieCrawlerProcedure(db)

	//TODO - Get ALL person
	personCrawlerProcedure(db)

}

func movieCrawlerProcedure(db *gorm.DB){
	genreAndMoviesAll(db)
	insertJSONsToDB("G:\\moviesData",db,"movie")
}

func personCrawlerProcedure(db *gorm.DB){
	//err := fetchPersonVisID(db)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	insertJSONsToDB("G:\\persons",db,"person")
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
		uris = append(uris,movie.Id)
	}

	webCrawler.FetchMovieInfosViaIDS(uris,db)

	return nil
}

func fetchPersonVisID(db *gorm.DB) error {
	uri := fileHost + peopleGZ
	var uris []int
	personData, err := GzFileDownloader.DownloadGZFile(uri)
	if err != nil {
		log.Println(err)
		return err
	}

	for _,person := range *personData{
		uris = append(uris,person.Id)
	}
	webCrawler.FetchPersonInfosViaIDS(uris,db)
	return nil
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

	//TODO - Insert Data to Database
	_ ,err := webCrawler.GenreTableCreate(apiURL,db)
	if err != nil{
		log.Fatalln(err)
		return
	}

	fetchMovieViaID(db)
	//making a function to handle fetching movies for genre
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

func insertJSONsToDB(dirPath string,db *gorm.DB,jsonType string){
	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return
	}

	if jsonType == "movie"{
		for _,file := range dir{
			err := movieJsonToDB(db, dirPath, file.Name())
			if err != nil {
				log.Fatalln(err)
			}
		}
	}else if jsonType == "person"{
		for _,file := range dir{
			err := personJsonToDB(db,dirPath,file.Name())
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

}

func movieJsonToDB(db *gorm.DB,dirPath string,fileName string)error{
	var movieInfo webCrawler.MovieInfo
	location := fmt.Sprintf("%s/%s",dirPath,fileName)
	jsonsData, err := ioutil.ReadFile(location)
	if err != nil {
		log.Println(err)
		return err
	}

	err = json.Unmarshal(jsonsData, &movieInfo)
	if err != nil {
		log.Println(err)
		return err
	}

	str, err := json.MarshalIndent(&movieInfo,"","\t")
	if err != nil {
		return err
	}

	ioutil.WriteFile(location,str,0666)

	if err := db.Where("id = ?",movieInfo.Id).First(&webCrawler.MovieInfo{});err !=nil{
		if errors.Is(err.Error,gorm.ErrRecordNotFound){
			//not found the record
			//insert to db
			db.Create(&movieInfo)
		}else{
			fmt.Println("???")
		}
	}
	return nil
}

func personJsonToDB(db *gorm.DB,dirPath string,fileName string) error {
	var personInfo webCrawler.PersonInfo
	location := fmt.Sprintf("%s/%s",dirPath,fileName)

	jsonData, err := ioutil.ReadFile(location)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonData, &personInfo)
	if err != nil {
		return err
	}

	if personInfo.ProfilePath == "" || len(personInfo.MovieCredits.Cast) == 0 && len(personInfo.MovieCredits.Crew) == 0{
		fmt.Println(personInfo.Name)
		return nil
	}



	if dbErr := db.Where("id = ?",personInfo.Id).First(&webCrawler.PersonInfo{});dbErr != nil{
		if errors.Is(dbErr.Error,gorm.ErrRecordNotFound){
			//TODO - ForEach cast need to check the movie info is our
			var newMovieCast []webCrawler.MovieCharacter
			var newMovieCrew []webCrawler.PersonCrew

			for _,castData := range personInfo.MovieCredits.Cast{
				//if current cast movie is existed
				if dbInsertErr := db.Where("id = ?",castData.MovieID).First(&webCrawler.MovieInfo{});dbInsertErr!=nil{
					if !errors.Is(dbInsertErr.Error,gorm.ErrRecordNotFound){
						//existed
						newMovieCast = append(newMovieCast,castData)
					}
				}
			}

			for _,crewData := range personInfo.MovieCredits.Crew{
				if dbInsertErr := db.Where("id = ?",crewData.MovieID).First(&webCrawler.MovieInfo{});dbInsertErr!=nil{
					if !errors.Is(dbInsertErr.Error,gorm.ErrRecordNotFound){
						//existed
						newMovieCrew = append(newMovieCrew,crewData)
					}
				}
			}

			personInfo.MovieCharacter = newMovieCast
			personInfo.PersonCrew = newMovieCrew
			db.Create(&personInfo)
		}
	}

	return nil
}
