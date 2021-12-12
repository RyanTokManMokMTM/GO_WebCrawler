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
	Iso6391     string `json:"iso_639_1"`
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
	RunTime          int     `json:"runtime"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`

	VideoInfos VideoResults `json:"videos" gorm:"-"`

	////gorm protocol
	//CreatedAt time.Time      `json:"-"`
	//UpdatedAt time.Time      `json:"-"`
	//DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	//Here have many2many relationship
	//one movie can have many genres
	//a genres can belong to many result

	GenreInfo  []GenreInfo      `json:"genres" gorm:"many2many:genres_movies"` //json do not contain this info, ignore that
	MovieVideo []MovieVideoInfo `json:"-" gorm:"foreignKey:MovieID"`
}

type MovieVideoInfo struct {
	Id       int    `gorm:"primarykey"`
	MovieID  uint   `gorm:"not null"`
	FilePath string `gorm:"not null"`
}

//GenreInfo TODO - Genre data
type GenreInfo struct {
	//APIResponse `gorm:"-"` //this info is no need in db

	//genre info
	Id   uint   `json:"id" gorm:"primarykey"`
	Name string `json:"name"`

	////gorm protocol
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type GenresMovies struct {
	Id          int `gorm:"primarykey"`
	MovieInfoId int `gorm:"autoIncrement:false"`
	GenreInfoId int `gorm:"autoIncrement:false"`
}

//PersonInfo TODO - Person data
type PersonInfo struct {
	Adult bool `json:"adult"`
	//also known as???
	Gender int  `json:"gender"` //1 or 2
	Id     uint `json:"id" gorm:"primarykey"`

	Department  string  `json:"known_for_department"`
	Name        string  `json:"name"`
	Popularity  float64 `json:"popularity"`
	ProfilePath string  `json:"profile_path"`
	//
	//CreatedAt time.Time      `json:"-"`
	//UpdatedAt time.Time      `json:"-"`
	//DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	//json only
	MovieCredits movieCreditAPIData `json:"movie_credits" gorm:"-"`
	//People has many movie character
	MovieCharacter []MovieCharacter `json:"-" gorm:"foreignKey:PersonID"`
	PersonCrew     []PersonCrew     `json:"-" gorm:"foreignKey:PersonID"`
}

type movieCreditAPIData struct {
	Cast []MovieCharacter `json:"cast"`
	Crew []PersonCrew     `json:"crew"`
}

type KnowFor struct {
	Adult bool `json:"adult"`
	//also known as???
	Gender int  `json:"gender"` //1 or 2
	Id     uint `json:"id" gorm:"primarykey"`

	Department  string  `json:"known_for_department"`
	Name        string  `json:"name"`
	Popularity  float64 `json:"popularity"`
	ProfilePath string  `json:"profile_path"`

	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	//People has many movie character
	MovieCharacter []MovieCharacter `json:"-" gorm:"foreignKey:PersonID"`
	PersonCrew     []PersonCrew     `json:"-" gorm:"foreignKey:PersonID"`
}

type MovieCharacter struct {
	//this data structure is about the person that what role of the movie is working for and some information
	//may be an actor? writer? a director? etc...

	//is a foreign key to  person
	//belong to
	PersonID uint `json:"-"` //current info belong to the user

	//belong to movie relationship
	MovieID   int       `json:"id"`
	MovieInfo MovieInfo `json:"-" gorm:"foreignKey:MovieID"`

	Id        uint   `json:"-" gorm:"primarykey"`
	Character string `json:"character"`
	CreditId  string `json:"credit_id"`
	Order     int    `json:"order"` // for current movie character order start:0

}

type PersonCrew struct {
	PersonID uint `json:"-"` //current info belong to the user

	//belong to movie relationship
	MovieID   int       `json:"id"`
	MovieInfo MovieInfo `json:"-" gorm:"foreignKey:MovieID"`

	Id         uint   `json:"-" gorm:"primarykey"`
	CreditId   string `json:"credit_id"`
	Department string `json:"department"`
}
