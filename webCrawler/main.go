package webCrawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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

const (
	host string = "https://api.themoviedb.org/3"
	apiKey string = "29570e7acc52b3e085ab46f6a60f0a55"
)

//TODO - GETTING API BASE INFO RESPONSE

type APIResponse struct{
	Page int `json:"page"`
	TotalPages int `json:"total_pages"`
}

//MOVIE AND GENRE RESPONSE
type movieAPIResponse struct {
	APIResponse
	Movies []MovieInfo `json:"results"`
}

type genreAPIResponse struct {
	Genres []GenreInfo `json:"genres"`
}

type peopleAPIResponse struct {
	APIResponse
	PeopleInfo []PersonInfo `json:"results"`
}

type peopleMovieCreditsAPIResponse struct {
	Cast []KnowFor `json:"cast"` //movie character
	Crew []KnowFor `json:"crew"` //movie crew director....
}

type creditTypeAPIResponse struct {
	//CreditType string `json:"credit_type"`
	//Department string `json:"department"`
	Job        string `json:"job"`
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

//PersonInfo TODO - Person data
type PersonInfo struct {
	Adult  bool `json:"adult"`
	Gender int  `json:"gender"`
	Id     uint `json:"id" gorm:"primarykey"`

	//getting All Know From another api...
	//ignore it ...
	//KnownFor []KnowFor `json:"known_for" gorm:"-""`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`

	//Job        string `json:"job"`


	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	//People has many movie character
	KnowFors []KnowFor `json:"-" gorm:"foreignKey:PersonID"`
}

type KnowFor struct {
	//this data structure is about the person that what role of the movie is working for and some information
	//may be an actor? writer? a director? etc...

	//is a foreign key to  person
	//belong to
	PersonID uint   `json:"-"`  //current info belong to the user

	//belong to movie relationship
	MovieID   int       `json:"id"`
	MovieInfo MovieInfo `json:"-" gorm:"foreignKey:MovieID"`

	Id        uint   `json:"-" gorm:"primarykey"`
	Character string `json:"character"`
	CreditId  string `json:"credit_id"`
	Order     int    `json:"order"` // for current movie character order start:0

	//Department string `json:"department"` //only crew but cast get from credit api
	Job        string `json:"job"` //only crew but cast get from credit api,current movie job
}

// GenreTableCreate TODO - Getting total page of the API response
func GenreTableCreate(uri string,db *gorm.DB) ([]GenreInfo, error){
	request, err := http.NewRequest("GET",uri,nil)
	if err != nil{
		log.Println(err)
		return nil,err
	}
	if err != nil {
		log.Println(err)
		return nil,err
	}

	res, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	defer res.Body.Close()
	fmt.Println(res.Body)
	body , err := ioutil.ReadAll(res.Body)
	fmt.Println(len(body))
	if err != nil{
		log.Println(err)
		return nil,err
	}

	var genres genreAPIResponse
	err = json.Unmarshal(body, &genres)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	//fmt.Println(len(genres.Genres))
	// TODO - before inset into database ,need to translate some text to traditional chinese : USE OPEN-CC HERE


	var dbGenres []GenreInfo = genres.Genres
	db.Create(&dbGenres)
	//
	for _, genre := range genres.Genres{
		fmt.Println(genre.Name)
	}
	return dbGenres,nil
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
func FetchMovieInfos(uris []string,db *gorm.DB,dataType string) bool{
	for _,uri := range uris{
		if dataType == "movie"{
			getMovieFromUri(uri,db) //try it first
		}else if dataType == "genre"{
			getMoviesGenres(uri,db)
		} else if dataType == "people" {
			getPeopleFromUri(uri,db)
		}
	}
	//fmt.Println(uris[0])
	//getPeopleFromUri(uris[0],db)
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
	for _,movie := range movieRes.Movies{
		//need to check current movie is in db?
		if movie.Overview == ""{
			continue
		}
		if dbErr := db.Where("id = ?",movie.Id).First(&MovieInfo{}); dbErr != nil{
			if errors.Is(dbErr.Error,gorm.ErrRecordNotFound) {
				var currentMovie MovieInfo = movie
				var genreIds []int = movie.GenreIds
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
				fmt.Printf("%v movie is inserted",movie.Title)
			}else{
				fmt.Println(movie.Title,"is existed")
			}
		}
	}
}

func getMoviesGenres(uri string,db *gorm.DB){
	// https://api.themoviedb.org/3/discover/movie?api_key=29570e7acc52b3e085ab46f6a60f0a55&language=zh-TW&sort_by=popularity.desc&page=1&with_genres=28&with_watch_monetization_types=flatrate
	//fetechingURI := host + allMovieURI + "?api_key=" + apiKey + "&language=zh-TW&sort_by=popularity.desc&page=1&with_genres="+strconv.Itoa(int(genreID))+"&with_watch_monetization_types=flatrate"
	var movieRes movieAPIResponse
	req,err := http.NewRequest("GET",uri,nil)
	if err != nil {
		log.Println(err)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	//read data from body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return
	}

	err = json.Unmarshal(body, &movieRes)
	if err != nil {
		log.Println(err)
		return
	}

	for _,movie := range movieRes.Movies{
		//need to check current movie is in db?
		if movie.Overview == ""{
			continue
		}
		if dbErr := db.Where("id = ?",movie.Id).First(&MovieInfo{}); dbErr != nil{
			if errors.Is(dbErr.Error,gorm.ErrRecordNotFound) {
				var currentMovie MovieInfo = movie
				var genreIds []int = movie.GenreIds
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
				fmt.Printf("%v movie is inserted",movie.Title)
			}
		}
	}
}

/*
	TODO - GETTING ALL NEEDED DATA -> at least 3 request!
	TODO - GETTING BASIC INFO FOR THE PEOPLE(Department: Acting or Directing only)
	TODO(API)
		-GETTING ALL PEOPLE -/person/popular
			-Return a []PersonInfo
		-GETTING ALL MOVIES CREW FOR CURRENT PEOPLE - /person/{person_id}/movie_credits
			-Return a []KnowFor
		-GETTING CURRENT PEOPLE specific job with its `credit_id` - /credit/{credit_id}
			-Return a credit info
		- Combine all data and insert to database
*/


//getActorFromUri TODO - fetching data from uris list(people)
func getPeopleFromUri(uri string,db *gorm.DB){
	var peopleRes peopleAPIResponse
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

	err = json.Unmarshal(body,&peopleRes)
	if err != nil {
		log.Fatalln(err)
	}

	for _,person := range peopleRes.PeopleInfo{
		knowForList,err := getPeopleKnowFor(person.Id,db) //getting all cast list for this people
		if err != nil {
			log.Fatalln(err)
		}

		//may not need this info...
		person.KnowFors = knowForList //may be empty
		indent, err := json.MarshalIndent(person, "", "\t")
		if err != nil {
			return
		}

		fmt.Println(string(indent))
		db.Create(&person)
	}

	//for each list has/have a group of genre -> separated it!
}

func getPeopleKnowFor(personID uint,db *gorm.DB) ([]KnowFor,error){
	//convert int to string
	var peopleCreditRes peopleMovieCreditsAPIResponse
	var result []KnowFor //include all cast and crew for current people

	personDataURI := host + "/person/" + strconv.Itoa(int(personID)) + "/movie_credits?api_key=" + apiKey + "&language=zh-TW"
	req ,err := http.NewRequest("GET",personDataURI,nil)
	if err != nil{
		log.Println(err)
		return nil,err
	}

	res, err := client.Do(req)
	if err != nil{
		log.Println(err)
		return nil,err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil{
		log.Println(err)
		return nil,err
	}

	err = json.Unmarshal(body, &peopleCreditRes)
	if err != nil {
		log.Println(err)
		return nil,err
	}

	//Cast
	if len(peopleCreditRes.Cast) > 0{
		for i:=0 ; i<len(peopleCreditRes.Cast);i++{
		//	var movie MovieInfo
			//here we need to check the movie db
			if movieErr := db.Where("id = ?", peopleCreditRes.Cast[i].MovieID).First(&MovieInfo{});movieErr != nil{
				if errors.Is(movieErr.Error,gorm.ErrRecordNotFound){
					continue
				}
			}
			//if current movie is exited
			//getting current person extra data include job department from api
			//then set it to our data struct and append
			creditData, err := getPeopleCredit(peopleCreditRes.Cast[i].CreditId)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			peopleCreditRes.Cast[i].Job = creditData.Job
			//peopleCreditRes.Cast[i].Department = creditData.Department
			peopleCreditRes.Cast[i].PersonID = personID
			result = append(result,peopleCreditRes.Cast[i])
		}
	}

	//Crew
	if len(peopleCreditRes.Crew) > 0 {
		for i:=0 ;i<len(peopleCreditRes.Crew);i++{
			//determine department = 'directing' and job = 'Director' only
			//ignore other....
			//here we need to check the movie db
			if movieErr := db.Where("id = ?", peopleCreditRes.Crew[i].MovieID).First(&MovieInfo{});movieErr != nil{
				if errors.Is(movieErr.Error,gorm.ErrRecordNotFound){
					continue
				}
			}

			if peopleCreditRes.Crew[i].Job != "Director"{
				continue
			}

			//current data belong to current person
			peopleCreditRes.Crew[i].PersonID = personID
			result = append(result,peopleCreditRes.Crew[i])
		}
	}

	return result,nil
}

func getPeopleCredit(creditID string) (*creditTypeAPIResponse ,error){
	//sending request to
	var creditRes creditTypeAPIResponse
	creditUri := host + "/credit/" + creditID + "?api_key=" +apiKey

	req,err := http.NewRequest("GET",creditUri,nil)
	if err != nil{
		log.Println(err)
		return nil,err
	}

	res,err := client.Do(req)
	if err != nil{
		log.Println(err)
		return nil,err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil,err
	}

	err = json.Unmarshal(body, &creditRes)
	if err != nil {
		log.Println(err)
		return nil,err
	}
	return &creditRes,nil
}
