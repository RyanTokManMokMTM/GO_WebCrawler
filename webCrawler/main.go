package webCrawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

/*
	TODO - Model relationship between movie and genre
*/


var client *http.Client

func init(){
	client = &http.Client{}

}

var (
	FetchERRFiled = errors.New("Fetch Err Failed ")
)


//TODO - GETTING API BASE INFO RESPONSE
type APIResponse struct{
	Page int `json:"page"`
	TotalPages int `json:"total_pages"`
}

type movieAPIResponse struct {
	APIResponse
	Movies []MovieInfo `json:"results"`
}

type genreAPIResponse struct {
	Genres []GenreInfo `json:"genres"`
}

// TODO - Database schema

//MovieInfo TODO - GETTING DATA FROM API
type MovieInfo struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIds         []int   `json:"genre_ids" gorm:"-"` //we are going to store it with join table ,ignore that...
	Id               uint    `json:"id" gorm:"primarykey"`
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
	
	//gorm protocol
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	//Here have many2many relationship
	//one movie can have many genres
	//a genres can belong to many result

	GenreInfo []GenreInfo `json:"-" gorm:"many2many:genres_movies"` //json do not contain this info, ignore that
}

//GenreInfo TODO - Genre data
type GenreInfo struct {
	//APIResponse `gorm:"-"` //this info is no need in db

	//genre info
	Id uint `json:"id" gorm:"primarykey"`
	Name string `json:"name"`

	////gorm protocol
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func GenreTableCreate(uri string,db *gorm.DB) error{
	request, err := http.NewRequest("GET",uri,nil)
	if err != nil{
		log.Println(err)
		return err
	}
	if err != nil {
		log.Println(err)
		return err
	}

	res, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return err
	}
	defer res.Body.Close()
	fmt.Println(res.Body)
	body , err := ioutil.ReadAll(res.Body)
	fmt.Println(len(body))
	if err != nil{
		log.Println(err)
		return err
	}

	var genres genreAPIResponse
	err = json.Unmarshal(body, &genres)
	if err != nil {
		log.Println(err)
		return err
	}
	//fmt.Println(len(genres.Genres))
	// TODO - before inset into database ,need to translate some text to traditional chinese : USE OPEN-CC HERE


	var dbGenres []GenreInfo = genres.Genres
	db.Create(&dbGenres)
	//
	for _, genre := range genres.Genres{
		fmt.Println(genre.Name)
	}
	return nil
}

// FetchPageInfo TODO - just fetching basic information that server is needed...
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

//FetchMovieInfos TODO - fetching data from uris list
func FetchMovieInfos(uris []string,db *gorm.DB) bool{
	for _,uri := range uris{
		getMovieFromUri(uri,db) //try it first
	}
	return true
}

//getMovieFromUri TODO - getting data from the specific URI
func getMovieFromUri(uri string,db *gorm.DB){
	var movieRes movieAPIResponse
	request , err := http.NewRequest("GET",uri,nil)
	if err != nil {
		log.Println(err)
		return
	}

	res, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	err = json.Unmarshal(body,&movieRes)
	if err != nil {
		log.Fatalln(err)
	}

	//for each list has/have a group of genre -> separated it!
	fmt.Println()
	for _,moviesInfo := range movieRes.Movies{
		var currentMovie MovieInfo = moviesInfo
		var genreIds []int = moviesInfo.GenreIds
		var genreList []GenreInfo
		for _,genreId := range genreIds{
			//for each genre list
			//let's test
			curGenre := GenreInfo{
				Id: uint(genreId),
			}
			genreList = append(genreList,curGenre)
		}
		currentMovie.GenreInfo = genreList
		db.Create(&currentMovie)
		fmt.Println(currentMovie)
	}
}
