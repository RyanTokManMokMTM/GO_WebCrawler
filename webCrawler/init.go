package webCrawler

import "net/http"

var client *http.Client

func init() {
	client = &http.Client{}
}

const (
	host       string = "https://api.themoviedb.org/3"
	apiKey     string = "29570e7acc52b3e085ab46f6a60f0a55"
	maxRoutine        = 90
)

var (
	detailURI = `%s/movie/%d?api_key=%s&language=zh-TW&append_to_response=videos&include_video_language=en`
)
