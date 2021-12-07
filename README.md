### TMDB Crawler
This application is used for fetching all movies info and crews info from TMDB api.

---
#### Package

**GzFileDownloader**
`(Download jsonGz files from TMDB)`
> (Movies):http://files.tmdb.org/p/exports/movie_ids_MM_DD_YYYY.json.gz.json.gz
> (People):http://files.tmdb.org/p/exports/person_ids_MM_DD_YYYY.json.gz

Function:
```go
`Arg: a string of GzFile URL`
`Return: A list of JSON Id and a error(if any)`
func DownloadGZFile(url string) (*[]*TMDBJson,error)
```
```go
type TMDBJson struct {
	Id int `json:"id"`
}
```
  
**webCrawler** 
`Fetch Movies Info and Related Persons Info from TMDB`  
**Movie and Person Data**
```go
// Movies Sturct
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
	RunTime 		 int 	  `json:"runtime"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	
	////gorm protocol
	//CreatedAt time.Time      `json:"-"`
	//UpdatedAt time.Time      `json:"-"`
	//DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	//Here have many2many relationship
	//one movie can have many genres
	//a genres can belong to many result

	GenreInfo []GenreInfo `json:"genres" gorm:"many2many:genres_movies"` //json do not contain this info, ignore that
}
```
```go
//PersonInfo Struct
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
```

Function Usage
`All Crawlering using Concurrcy to improve the performance`  
> Movie Fetcher
```go
@Parms: ids : a list of movie ids
@Parms: moviePath : json data to store at some location
func FetchMovieInfosViaIDS(ids []int,moviePath string)
```
> Person Fetcher
```go
@Parms: ids : a list of person ids
@Parms: personPath : json data to store at some location
func FetchPersonInfosViaIDS(ids []int,personPath string)
```

