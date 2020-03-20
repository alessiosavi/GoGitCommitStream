package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	stringutils "github.com/alessiosavi/GoGPUtils/string"
	"github.com/alessiosavi/GoGitCommitStream/structure"
)

func core(hour, minutes, second int, url string) bool {
	var (
		resp         *http.Response
		err          error
		body         string
		notification []structure.GitStream
	)

	if resp, err = http.Get(url); err != nil {
		fmt.Println(err)
		return false
	}
	if body, err = getBody(resp.Body); err != nil {
		fmt.Println(err)
		return false
	}
	if err = json.Unmarshal([]byte(body), &notification); err != nil {
		fmt.Println(err)
		return false
	}
	if len(notification) < 1 {
		fmt.Println("Not enough data")
		return false
	}

	t := time.Now()
	currentDate := time.Date(t.Year(), t.Month(), t.Day(), hour, minutes, second, 0, time.Local)
	n := notification[0]
	return n.Commit.Author.Date.After(currentDate)
}

type server struct {
	hours   int
	minutes int
	second  int
	url     string
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"updated": "%t"}`, core(s.hours, s.minutes, s.second, s.url))))
}

func main() {
	githuUrl := flag.String("url", "", "url related to the github.com project")
	port := flag.Int("port", 8080, "port to spawn the server")
	hour := flag.Int("hour", 18, "hour related to the commit time to check")
	minutes := flag.Int("minutes", 0, "minutes related to the commit time to check")
	seconds := flag.Int("seconds", 0, "seconds related to the commit time to check")
	flag.Parse()

	if stringutils.IsBlank(*githuUrl) {
		panic("url is a mandatory parameter")
	}

	s := &server{hours: *hour, minutes: *minutes, second: *seconds, url: *githuUrl}
	http.Handle("/", s)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}

// GetBody is delegated to retrieve the body from the given response
func getBody(body io.ReadCloser) (string, error) {
	var sb strings.Builder
	var err error

	defer body.Close()
	if _, err = io.Copy(&sb, body); err != nil {
		return "", nil
	}
	return sb.String(), nil
}
