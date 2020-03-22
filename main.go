package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	stringutils "github.com/alessiosavi/GoGPUtils/string"
	"github.com/alessiosavi/GoGitCommitStream/structure"
	"github.com/aws/aws-lambda-go/lambda"
)

func core(hour, minutes, second int, url string) bool {
	var (
		resp         *http.Response
		err          error
		body         string
		notification []structure.GitStream
		loc          *time.Location
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
	if len(notification) == 0 {
		fmt.Println("Not enough data")
		return false
	}

	if loc, err = time.LoadLocation("Europe/Rome"); err != nil {
		panic(err)
	}
	time.Local = loc

	t := time.Now().Local()
	targetDate := time.Date(t.Year(), t.Month(), t.Day(), hour, minutes, second, 0, loc)
	fmt.Printf("Target date: %+v\n", targetDate)

	n := notification[0]
	fmt.Printf("Git date: %+v\n", n.Commit.Author.Date)
	fmt.Printf("Git date normalized: %+v\n", n.Commit.Author.Date.Local())
	return n.Commit.Author.Date.Local().After(targetDate)
}

type InputRequest struct {
	Hours   int    `json:"hours"`
	Minutes int    `json:"minutes"`
	Second  int    `json:"second"`
	Url     string `json:"url"`
}

func main() {
	lambda.Start(HandleRequest)
	//console()
}

func console() {
	url := flag.String("url", "", "url related to the github.com project")
	hour := flag.Int("hour", 18, "hour related to the commit time to check")
	minutes := flag.Int("minutes", 0, "minutes related to the commit time to check")
	seconds := flag.Int("seconds", 0, "seconds related to the commit time to check")
	flag.Parse()

	if stringutils.IsBlank(*url) {
		panic("url is a mandatory parameter")
	}

	fmt.Printf(`{"updated": "%t","datetime":"%s"}`, core(*hour, *minutes, *seconds, *url), time.Now().Format(time.RFC3339))
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

func HandleRequest(ctx context.Context, request InputRequest) (string, error) {
	return fmt.Sprintf(`{"updated": "%t","datetime":"%s"}`, core(request.Hours, request.Minutes, request.Second, request.Url), time.Now().Format(time.RFC3339)), nil
}
