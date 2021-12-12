package webCrawler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"strconv"
	"sync"
	"unicode"

	"gorm.io/gorm"
)

// GenreTableCreate TODO - Getting total page of the API response
func GenreTableCreate(uri string, db *gorm.DB) ([]GenreInfo, error) {
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := client.Do(request)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()
	fmt.Println(res.Body)
	body, err := ioutil.ReadAll(res.Body)
	fmt.Println(len(body))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var genres genreAPIResponse
	err = json.Unmarshal(body, &genres)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	//fmt.Println(len(genres.Genres))
	// TODO - before inset into database ,need to translate some text to traditional chinese : USE OPEN-CC HERE

	var dbGenres []GenreInfo = genres.Genres
	db.Create(&dbGenres)
	//
	for _, genre := range genres.Genres {
		fmt.Println(genre.Name)
	}
	return dbGenres, nil
}

// FetchPageInfo TODO - just fetching basic information that server is needed...
func FetchPageInfo(uri string) int {
	request, err := http.NewRequest("GET", uri, nil)
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
	//fmt.Println(res.Header)
	var result APIResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(result)
	return result.TotalPages //get the total page of current response
}

func FetchMovieInfosViaIDS(ids []int, moviePath string) {
	wg := sync.WaitGroup{}
	movieURIsCh := make(chan string, 100)
	fetchResultCh := make(chan *MovieInfo, 100) // all result are movie info

	//using a goroutine to print out the result
	go getMovieResultAndConvertTOJSON(fetchResultCh, moviePath) //this goroutine will wait the result

	go func() {
		for i := 0; i < maxRoutine; i++ {
			wg.Add(1)
			go asyncMovieFetcher(movieURIsCh, fetchResultCh, &wg) //used to fetch data ,if there is not any data need to fetch end!
		}
	}() //another goroutine

	//push all URIs to the channel
	for _, id := range ids {
		reqURI := fmt.Sprintf(detailURI, host, id, apiKey)
		movieURIsCh <- reqURI
	}
	fmt.Println("pushing data finished and closing the channel...")
	close(movieURIsCh)
	defer close(fetchResultCh)
	wg.Wait()
	fmt.Println("Fetching is Done....")
}

//httpGETData TODO - RETURN THE RESULT AND ERROR IF IT HAVE AN ERROR
func httpGETMovieData(uri string) *MovieInfo {
	var movieDetail movieDetailAPIResponse
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalln(err)
		return nil
	}

	err = json.Unmarshal(body, &movieDetail)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	if movieDetail.Overview == "" {
		return nil
	}

	//check title is chinese
	isChin := isChinese(movieDetail.Title)
	if !isChin {
		return nil
	}

	return &movieDetail.MovieInfo
}

func getMovieResultAndConvertTOJSON(result chan *MovieInfo, moviePath string) {
	for {
		v, ok := <-result
		if !ok { //getting nothing,channel closed
			//fmt.Println("Not any result!",ok)
			break
		}
		//fmt.Println
		if v != nil {
			fmt.Println(v.Title)
			toMovieJson(v, moviePath)
		}
	}
}

func toMovieJson(movie *MovieInfo, filePath string) {
	fileName := filePath + "/" + strconv.Itoa(int(movie.Id)) + ".json"
	f, err := os.Create(fileName)
	if err != nil {
		return
	}

	data, err := json.MarshalIndent(movie, "", "\t")
	if err != nil {
		return
	}

	f.Write(data)
	f.Close()
}

func asyncMovieFetcher(ids chan string, result chan *MovieInfo, wg *sync.WaitGroup) {
	defer (*wg).Done() // each goroutine

	for {
		v, ok := <-ids
		if !ok {
			//fmt.Println("read all data!",ok)
			break
		}
		result <- httpGETMovieData(v)

	}
}

// FetchPersonInfosViaIDS TODO - Fetch all person
func FetchPersonInfosViaIDS(ids []int, personPath string) {
	//2 channels
	wg := sync.WaitGroup{}
	personIdsCh := make(chan string, 100)
	resultCh := make(chan *PersonInfo, 100)

	//set a go routine to
	go getPersonResultAndConvertTOJSON(resultCh, personPath)

	go func() {
		for i := 0; i < maxRoutine; i++ {
			wg.Add(1)
			go asyncPersonFetcher(personIdsCh, resultCh, &wg)
		}
	}()

	for _, id := range ids {
		personURI := fmt.Sprintf("%s/person/%d?api_key=%s&language=zh-TW&append_to_response=movie_credits", host, id, apiKey)
		personIdsCh <- personURI
	}
	fmt.Println("pushing data finished and closing the channel...")
	close(personIdsCh) // change set
	defer close(resultCh)
	wg.Wait()
	fmt.Println("Fetching is Done....")
}

func getPersonResultAndConvertTOJSON(result chan *PersonInfo, personPath string) {
	for {
		v, ok := <-result
		if !ok {
			fmt.Println("channel is closed!")
			break
		}

		if v != nil {
			err := toPersonJSON(v, personPath)
			if err != nil {
				log.Fatalln(err)
			}
		}
		//get the result out and convert to json
		//toPersonJSON(v)
	}
}

func asyncPersonFetcher(uris chan string, result chan *PersonInfo, wg *sync.WaitGroup) {
	defer (*wg).Done()
	for {
		v, ok := <-uris
		if !ok {
			fmt.Println("channel is closed :", ok)
			break
		}

		//fetch the data get and push the data to the channel
		result <- httpGETPersonData(v)
	}
}

func toPersonJSON(person *PersonInfo, personPath string) error {
	fileName := fmt.Sprintf("%s/%d.json", personPath, person.Id)
	fmt.Println(person.Name)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := json.MarshalIndent(person, "", "\t")
	if err != nil {
		return err
	}
	file.Write(data)
	return nil
}

func httpGETPersonData(uri string) *PersonInfo {
	var personData PersonInfo
	res, err := http.Get(uri)
	if err != nil {
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

func isChinese(chinese string) bool {
	count := 0
	for _, v := range chinese {
		if unicode.Is(unicode.Han, v) {
			count++
			break
		}
	}

	return count > 0
}
