package webCrawler

import (
	"gorm.io/gorm"
	"time"
)

//TODO - GETTING API BASE INFO RESPONSE

type APIResponse struct {
	Page       int `json:"page"`
	TotalPages int `json:"total_pages"`
}

//MOVIE AND GENRE RESPONSE
type movieAPIResponse struct {
	APIResponse
	Movies []MovieInfo `json:"results"`
}

type G interface{}

type genreAPIResponse struct {
	Genres []GenreInfo `json:"genres"`
}

type movieDetailAPIResponse struct {
	MovieInfo
}

type VideoResults struct {
	Results []TMDBVideoInfo `json:"results"`
}

// TODO - Database schema

type TMDBVideoInfo struct {
	Iso6391     string `json:"iso_639_1" db:""`
	Iso31661    string `json:"iso_3166_1"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Site        string `json:"site"`
	Size        int    `json:"size"`
	Type        string `json:"type"`
	Official    bool   `json:"official"`
	PublishedAt string `json:"published_at"`
	Id          string `json:"id"`
}

//MovieInfo TODO - GETTING DATA FROM API -need chinese and chinese overview only
type MovieInfo struct {
	Adult            bool    `json:"adult" gorm:"not null"`
	BackdropPath     string  `json:"backdrop_path" gorm:"not null"`
	GenreIds         []int   `json:"-" gorm:"-" gorm:"not null"` //we are going to store it with join table ,ignore that...
	Id               uint    `json:"id" gorm:"primarykey" gorm:"not null"`
	OriginalLanguage string  `json:"original_language" gorm:"not null"`
	OriginalTitle    string  `json:"original_title" gorm:"not null"`
	Overview         string  `json:"overview" gorm:"not null"`
	Popularity       float64 `json:"popularity" gorm:"not null"`
	PosterPath       string  `json:"poster_path" gorm:"not null"`
	ReleaseDate      string  `json:"release_date" gorm:"not null"`
	Title            string  `json:"title" gorm:"not null"`
	RunTime          int     `json:"runtime" gorm:"not null"`
	Video            bool    `json:"video" gorm:"not null"`
	VoteAverage      float64 `json:"vote_average" gorm:"not null"`
	VoteCount        int     `json:"vote_count" gorm:"not null"`

	//VideoInfos VideoResults `json:"videos" gorm:"-"`

	////gorm protocol
	//CreatedAt time.Time      `json:"-"`
	//UpdatedAt time.Time      `json:"-"`
	//DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	//Here have many2many relationship
	//one movie can have many genres
	//a genres can belong to many result

	GenreInfo []GenreInfo `json:"genres" gorm:"many2many:genres_movies"` //json do not contain this info, ignore that
	//MovieVideo []MovieVideoInfo `json:"-" gorm:"foreignKey:Id"`

	CreatedAt time.Time      `json:"-" gorm:"type:timestamp"`
	UpdatedAt time.Time      `json:"-" gorm:"type:timestamp"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:timestamp" json:"-"`
}

//type MovieVideoInfo struct {
//	MovieID     uint   `gorm:"primarykey"`
//	FilePath    string `gorm:"primarykey"`
//	TrailerName string
//	ReleaseTime time.Time
//}

//GenreInfo TODO - Genre data
type GenreInfo struct {
	//APIResponse `gorm:"-"` //this info is no need in db

	//genre info
	GenreID   uint        `json:"id" gorm:"primarykey" gorm:"not null"`
	Name      string      `json:"name" gorm:"not null"`
	MovieInfo []MovieInfo `gorm:"many2many:genres_movies" json:"-"`
	////gorm protocol
	CreatedAt time.Time      `json:"-" gorm:"type:timestamp"`
	UpdatedAt time.Time      `json:"-" gorm:"type:timestamp"`
	DeletedAt gorm.DeletedAt `gorm:"index;type:timestamp" json:"-"`
}

//type GenresMovies struct {
//	Id          uint `gorm:"primarykey,not null"`
//	MovieInfoId int  `gorm:"autoIncrement:false,not null"`
//	GenreInfoId int  `gorm:"autoIncrement:false,not null"`
//}

//PersonInfo TODO - Person data
type PersonInfo struct {
	Adult bool `json:"adult" gorm:"not null"`
	//also known as???
	Gender   int  `json:"gender" gorm:"not null"` //1 or 2
	PersonID uint `json:"id" gorm:"primarykey" gorm:"not null"`

	Department  string  `json:"known_for_department" gorm:"not null"`
	Name        string  `json:"name" gorm:"not null"`
	Popularity  float64 `json:"popularity" gorm:"not null"`
	ProfilePath string  `json:"profile_path" gorm:"not null"`
	//
	//CreatedAt time.Time      `json:"-"`
	//UpdatedAt time.Time      `json:"-"`
	//DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	//json only
	MovieCredits movieCreditAPIData `json:"movie_credits" gorm:"-"`
	//People has many movie character
	MovieCharacter []MovieCharacter `json:"-" gorm:"foreignKey:PersonID" gorm:"not null"`
	PersonCrew     []PersonCrew     `json:"-" gorm:"foreignKey:PersonID" gorm:"not null"`
}

type movieCreditAPIData struct {
	Cast []MovieCharacter `json:"cast"`
	Crew []PersonCrew     `json:"crew"`
}

type KnowFor struct {
	Adult bool `json:"adult" gorm:"not null"`
	//also known as???
	Gender    int  `json:"gender" gorm:"not null"` //1 or 2
	KnowForID uint `json:"id" gorm:"primarykey" gorm:"not null"`

	Department  string  `json:"known_for_department" gorm:"not null"`
	Name        string  `json:"name" gorm:"not null"`
	Popularity  float64 `json:"popularity" gorm:"not null"`
	ProfilePath string  `json:"profile_path" gorm:"not null"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	//People has many movie character
	MovieCharacter []MovieCharacter `json:"-" gorm:"foreignKey:PersonID" gorm:"not null"`
	PersonCrew     []PersonCrew     `json:"-" gorm:"foreignKey:PersonID" gorm:"not null"`
}

type MovieCharacter struct {
	//this data structure is about the person that what role of the movie is working for and some information
	//may be an actor? writer? a director? etc...

	//is a foreign key to  person
	//belong to
	PersonID uint `json:"-" gorm:"not null"` //current info belong to the user

	//belong to movie relationship
	MovieID   int       `json:"id" gorm:"not null"`
	MovieInfo MovieInfo `json:"-" gorm:"foreignKey:Id" gorm:"not null"`

	Id        uint   `json:"-" gorm:"primarykey" gorm:"not null"`
	Character string `json:"character" gorm:"not null"`
	CreditId  string `json:"credit_id" gorm:"not null"`
	Order     int    `json:"order" gorm:"not null"` // for current movie character order start:0

}

type PersonCrew struct {
	PersonID uint `json:"-" gorm:"not null"` //current info belong to the user

	//belong to movie relationship
	MovieID   int       `json:"id" gorm:"not null"`
	MovieInfo MovieInfo `json:"-" gorm:"foreignKey:Id" gorm:"not null"`

	Id         uint   `json:"-" gorm:"primarykey" gorm:"not null"`
	CreditId   string `json:"credit_id" gorm:"not null"`
	Department string `json:"department" gorm:"not null"`
}
