package main

import (
	"fmt"
	"os"
	"time"
	"strings"
	"strconv"
	"regexp"
	"math"
	"sort"
)

func main() {
	args := os.Args[1:]

	requestCount := 1
	shouldProfile := false
	url := ""

	for i := 0; i < len(args); i++ {
		switch arg := args[i]; arg {
			case "--url":
				if len(args) < i+1 {
					fmt.Printf("Please specify a <URL> after --url flag")
					os.Exit(1)
				}
				url = strings.ToLower(args[i+1])
				isURL, _ := regexp.MatchString(`[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`, url)
				if isURL != true {
					fmt.Println("<URL> is malformed, please check to make sure it is standard")
					os.Exit(1)
				}
			case "--profile":
				if len(args) < i+1 {
					fmt.Printf("Please specify a <REQUEST_COUNT> after --profile flag")
					os.Exit(1)
				}
				count, err := strconv.Atoi(args[i+1])
				if err != nil {
					fmt.Printf("<REQUEST_COUNT> not a number")
					os.Exit(1)
				}
				requestCount = count
				shouldProfile = true

			case "--help":
				fmt.Println("")
				fmt.Printf("Basic Usage: go run . --url <URL>\n\n")
				fmt.Println("--url <URL> The URL that will be requested and print the body")
				fmt.Println("--profile <REQUEST_COUNT> Profile Requests with x count")
				return
		}
	}

	if url == "" {
		fmt.Println("Please specify a url via --url <URL>")
		return;
	}

	requests := make(chan ProfileData, requestCount)
	newURL, _ := parseURL(url)


	for i := 0; i < requestCount; i++  {
		go threadedRequest(newURL, requests)
	}
	

	profileData := make([]ProfileData, 0, requestCount)
	i := 0
	for request := range requests {

		profileData = append(profileData, request)
		i++
		if i >= requestCount {
			close(requests)
		}
	}

	if shouldProfile != true {
		fmt.Println("")
		fmt.Printf("---- Body Data for URL: " + url + " ----\n")
		fmt.Printf("%s", profileData[0].resp.body)
		return
	}


	times := make([]int, 0, requestCount)
	bytes := make([]int, 0, requestCount)
	errors := make([]int, 0, requestCount)
	sumTimes := 0
	for _, pData := range profileData {
		times = append(times, int(pData.timeTaken.Milliseconds()))
		sumTimes += int(pData.timeTaken.Milliseconds())
		if pData.resp.statusCode != 200 {
			errors = append(errors, pData.resp.statusCode)
		}
		bytes = append(bytes, pData.resp.respSize)
	}

	sort.Ints(times)
	sort.Ints(bytes)

	slowestTime := times[len(times)-1]
	fastestTime := times[0]
	
	meanTime := sumTimes / requestCount
	middle := int(math.Floor(float64(len(times))/2.0))
	medianTime := times[middle]
	
	smallestBytes := bytes[0]
	largestBytes := bytes[len(bytes)-1]

	requestsOK := float64(requestCount - len(errors))
	percentSuccess := requestsOK / float64(requestCount) * 100.0


	fmt.Println("")
	fmt.Printf("---- Profiling Data for URL: " + url + " ----\n")
	fmt.Printf("Number of Requests: %d\n", requestCount)
	fmt.Printf("Fastest Time: %d\n", fastestTime)
	fmt.Printf("Slowest Time: %d\n", slowestTime)
	fmt.Printf("Mean Time: %d | Median Time: %d\n", meanTime, medianTime)
	fmt.Printf("%% of Requests that succeded: %.1f%%\n", percentSuccess)

	if len(errors) > 0 {
		fmt.Printf("Error Codes not 200: %v\n", errors)
	} else {
		fmt.Printf("Error codes not 200: None\n")
	}
	

	fmt.Printf("Smallest byte response: %d\n", smallestBytes)
	fmt.Printf("Largest byte response: %d\n", largestBytes)
}

func threadedRequest(url URL, requests chan ProfileData) {
	profileData := ProfileData{}
	req := HTTPRequest{"GET", url.base, url.path}
	start := time.Now()
	data := sendRequest(req)
	
	since := time.Since(start)

	resp := parseResp(data)
	// Recursivly keep jumping points until a 400 or 200 status is given
	if resp.statusCode >= 300 && resp.statusCode < 400 {
		redirectUrl, _ := parseRedirect(resp.headers["Location"])
		go threadedRequest(redirectUrl, requests)
	} else {
		profileData.resp = resp
		profileData.timeTaken = since
	
		requests <- profileData
	}
}

type ProfileData struct {
	resp HTTPResponse 
	timeTaken time.Duration
}
