package webCrawler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"strconv"
	"sync"
	"time"
	"unicode"

	"gorm.io/gorm"
)

/*
	TODO - Model relationship between movie and genre
*/


var client *http.Client


func init(){
	client = &http.Client{}
	//ERR_NOT_CHINESE = errors.New("NOT INCLUDED CHINESE")
	//ERR_EMPTY_OVERVIEW = errors.New("EMPTY OVERVIEW")
}

const (
	host string = "https://api.themoviedb.org/3"
	apiKey string = "29570e7acc52b3e085ab46f6a60f0a55"
	maxRoutine = 90
)

var (
	detailURI = "%s/movie/%d?api_key=%s&language=zh-TW"
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

type movieDetailAPIResponse struct {
	MovieInfo
}

type DepartmentData struct {
	Department
	//Jobs []string `json:"jobs"`
}

// TODO - Database schema

//MovieInfo TODO - GETTING DATA FROM API -need chinese and chinese overview only
type MovieInfo struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIds         []int   `json:"-" gorm:"-"` //we are going to store it with join table ,ignore that...
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

	GenreInfo []GenreInfo `json:"genres" gorm:"many2many:genres_movies"` //json do not contain this info, ignore that
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

type GenresMovies struct {
	Id int `gorm:"primarykey"`
	MovieInfoId int `gorm:"autoIncrement:false"`
	GenreInfoId int `gorm:"autoIncrement:false"`
}

//PersonInfo TODO - Person data
type PersonInfo struct {
	Adult  bool `json:"adult"`
	//also known as???
	Gender int  `json:"gender"` //1 or 2
	Id     uint `json:"id" gorm:"primarykey"`

	Department string  `json:"known_for_department"`
	Name               string  `json:"name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
	//
	//CreatedAt time.Time      `json:"-"`
	//UpdatedAt time.Time      `json:"-"`
	//DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	//json only
	MovieCredits movieCreditAPIData `json:"movie_credits" gorm:"-"`
	//People has many movie character
	MovieCharacter []MovieCharacter `json:"-" gorm:"foreignKey:PersonID"`
	PersonCrew []PersonCrew `json:"-" gorm:"foreignKey:PersonID"`
}

type movieCreditAPIData struct{
	Cast []MovieCharacter `json:"cast"`
	Crew []PersonCrew `json:"crew"`
}

type KnowFor struct {
	Adult  bool `json:"adult"`
	//also known as???
	Gender int  `json:"gender"` //1 or 2
	Id     uint `json:"id" gorm:"primarykey"`

	Department string  `json:"known_for_department"`
	Name               string  `json:"name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	//People has many movie character
	MovieCharacter []MovieCharacter `json:"-" gorm:"foreignKey:PersonID"`
	PersonCrew []PersonCrew `json:"-" gorm:"foreignKey:PersonID"`
}

type MovieCharacter struct {
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

}

type PersonCrew struct {
	PersonID uint   `json:"-"`  //current info belong to the user

	//belong to movie relationship
	MovieID   int       `json:"id"`
	MovieInfo MovieInfo `json:"-" gorm:"foreignKey:MovieID"`

	Id        uint   `json:"-" gorm:"primarykey"`
	CreditId  string `json:"credit_id"`
	Department string `json:"department"`
}

type Department struct {
	Id int `gorm:"primarykey" json:"-"`
	DepartmentName string `json:"department" gorm:"primarykey" `

	//has a list of job
	//DepartmentJob []DepartmentJob `json:"-" gorm:"foreignKey:Id"`
}


//func (model *GenresMovies)AfterCreate(db *gorm.DB) (err error){
//
//	if err := db.Exec("ALTER TABLE genres_movies DROP CONSTRAINT genres_movies_pkey").Error ; err != nil {
//		log.Println(err)
//		return err
//	}
//
//	if err := db.Exec("ALTER TABLE genres_movies ADD CONSTRAINT  genres_movies_unique UNIQUE(genre_info_id,movie_info_id)").Error; err != nil{
//		log.Println(err)
//		return err
//	}
//
//	if err := db.Exec("ALTER TABLE genres_movies ADD CONSTRAINT genres_movies_pkey PRIMARY KEY (id)").Error ; err != nil{
//		log.Println(err)
//		return err
//	}
//	fmt.Println("Done")
//	return nil
//}


//type DepartmentJob struct {
//	Id int `gorm:"primarykey" json:"-"`
//	Job string `gorm:"primarykey" `
//}

//FetchAllDepartment TODO - Get All Department job -Get department only
func FetchAllDepartment(uri string,db *gorm.DB){
	fmt.Println(uri)
	var response []DepartmentData
	res, err := http.Get(uri)
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

	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Println(err)
		return
	}


	for _,department := range response{
		fmt.Println(department.DepartmentName)
		if department.DepartmentName == "Actors"{
			department.DepartmentName = "Acting"
		}
		if errDB := db.Where("department_name = ?",department.DepartmentName).First(&Department{});errDB != nil{
			if errors.Is(errDB.Error,gorm.ErrRecordNotFound){
				db.Create(&department.Department)
			}
		}
	}

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

//FetchMovieInfos TODO - fetching data from uris  - NOT USED
func FetchMovieInfos(uris []string,db *gorm.DB,dataType string) bool{
	for _,uri := range uris{
		if dataType == "movie"{
			getMovieFromUri(uri,db) //try it first
		}else if dataType == "genre"{
			getMoviesGenres(uri,db)
		}
		//} else if dataType == "people" {
		//	getPeopleFromUri(uri,db)
		//}

	}
	return true
}

func FetchMovieInfosViaIDS(ids []int,db *gorm.DB) {
	wg := sync.WaitGroup{}
	movieURIsCh := make(chan string,100)
	fetchResultCh := make(chan *MovieInfo,100)// all result are movie info

	//using a goroutine to print out the result
	go getMovieResultAndConvertTOJSON(fetchResultCh) //this goroutine will wait the result

	go func(){
		for i := 0;i<maxRoutine;i++{
			wg.Add(1)
			go asyncMovieFetcher(movieURIsCh,fetchResultCh,&wg) //used to fetch data ,if there is not any data need to fetch end!
		}
	}() //another goroutine

	//push all URIs to the channel
	for _,id := range ids{
		reqURI := fmt.Sprintf(detailURI,host,id,apiKey)
		movieURIsCh <- reqURI
	}
	fmt.Println("pushing data finished and closing the channel...")
	close(movieURIsCh)
	defer close(fetchResultCh)
	wg.Wait()
	fmt.Println("Fetching is Done....")
}

//httpGETData TODO - RETURN THE RESULT AND ERROR IF IT HAVE AN ERROR
func httpGETMovieData(uri string) *MovieInfo{
	var movieDetail movieDetailAPIResponse
	req,err := http.NewRequest("GET",uri,nil)
	if err != nil{
		log.Fatalln(err)
		return nil
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	defer res.Body.Close()

	body,err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	err = json.Unmarshal(body,&movieDetail)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	if movieDetail.Overview == ""{
		return nil
	}

	//check title is chinese
	isChin := isChinese(movieDetail.Title)
	if !isChin{
		return nil
	}

	return &movieDetail.MovieInfo
}

func getMovieResultAndConvertTOJSON(result chan *MovieInfo){
	for{
		v,ok := <- result
		if !ok{ //getting nothing,channel closed
			//fmt.Println("Not any result!",ok)
			break
		}
		//fmt.Println
		if v != nil{
			fmt.Println(v.Title)
			toMovieJson(v)
		}
	}
}

func toMovieJson(movie *MovieInfo) {
	fileName := "G:/moviesData/"+ strconv.Itoa(int(movie.Id)) + ".json"
	f, err := os.Create(fileName)
	if err != nil {
		return
	}

	data, err := json.MarshalIndent(movie,"","\t")
	if err != nil {
		return
	}

	f.Write(data)
	f.Close()
}

func asyncMovieFetcher(ids chan string,result chan *MovieInfo,wg *sync.WaitGroup){
	defer (*wg).Done() // each goroutine

	for{
		v,ok := <- ids
		if !ok{
			//fmt.Println("read all data!",ok)
			break
		}
		result <- httpGETMovieData(v)

	}
}

// FetchPersonInfosViaIDS TODO - Fetch all person
func FetchPersonInfosViaIDS(ids []int,db *gorm.DB){
	//2 channels
	wg := sync.WaitGroup{}
	personIdsCh := make(chan string,100)
	resultCh := make(chan *PersonInfo,100)

	//set a go routine to
	go getPersonResultAndConvertTOJSON(resultCh)

	go func(){
		for i:= 0; i< maxRoutine;i++{
			wg.Add(1)
			go asyncPersonFetcher(personIdsCh,resultCh,&wg)
		}
	}()

	for _,id := range ids{
		personURI := fmt.Sprintf("%s/person/%d?api_key=%s&language=zh-TW&append_to_response=movie_credits",host,id,apiKey)
		personIdsCh <- personURI
	}
	fmt.Println("pushing data finished and closing the channel...")
	close(personIdsCh) // change set
	defer close(resultCh)
	wg.Wait()
	fmt.Println("Fetching is Done....")
}

func getPersonResultAndConvertTOJSON(result chan *PersonInfo){
	for{
		v, ok := <- result
		if !ok {
			fmt.Println("channel is closed!")
			break
		}

		if v != nil{
			err := toPersonJSON(v)
			if err != nil {
				log.Fatalln(err)
			}
		}
		//get the result out and convert to json
		//toPersonJSON(v)
	}
}

func asyncPersonFetcher(uris chan string,result chan *PersonInfo,wg *sync.WaitGroup){
	defer (*wg).Done()
	for{
		v,ok := <- uris
		if !ok {
			fmt.Println("channel is closed :",ok)
			break
		}

		//fetch the data get and push the data to the channel
		result <- httpGETPersonData(v)
	}
}

func toPersonJSON(person *PersonInfo) error {
	fileName := fmt.Sprintf("G:/persons/%d.json",person.Id)
	fmt.Println(person.Name)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := json.MarshalIndent(person,"","\t")
	if err != nil {
		return err
	}
	file.Write(data)
	return nil
}

func httpGETPersonData(uri string) *PersonInfo{
	var personData PersonInfo
	res,err := http.Get(uri)
	if err != nil{
		fmt.Println(err)
		return nil
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return nil
	}

	err = json.Unmarshal(body, &personData)
	if err != nil {
		log.Println(err)
		return nil
	}
	
	return &personData
}

func getMovieDetail(uri string,db *gorm.DB) {
	var movieDetail movieDetailAPIResponse
	req,err := http.NewRequest("GET",uri,nil)
	if err != nil{
		log.Println(err)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer res.Body.Close()

	body,err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(res)
		return
	}

	err = json.Unmarshal(body,&movieDetail)
	if err != nil {
		log.Println(res)
		return
	}

		//need to check current movie is in db?
	if movieDetail.Overview != ""{
		if dbErr := db.Where("id = ?",movieDetail.Id).First(&MovieInfo{}); dbErr != nil{
			if errors.Is(dbErr.Error,gorm.ErrRecordNotFound) {
				db.Create(&movieDetail.MovieInfo)
				fmt.Printf("%v movie is inserted",movieDetail.MovieInfo.Id)
			}
		}
	}


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

func isChinese(chinese string) bool{
	count := 0
	for _,v := range chinese{
		if unicode.Is(unicode.Han,v){
			count++
			break
		}
	}

	return count > 0
}