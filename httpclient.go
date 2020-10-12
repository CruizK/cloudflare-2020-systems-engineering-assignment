package main


import (
	"fmt"
	"crypto/tls"
	"io/ioutil"
	"net"
	"bufio"
	"strings"
	"os"
	"regexp"
	"strconv"
	"time"
)

func sendRequest(req HTTPRequest) []byte {
	timeout, _ := time.ParseDuration("10s")
	d := net.Dialer{
		Timeout: timeout,
	}
	tlsConn, err := tls.DialWithDialer(&d, "tcp", req.host + ":https", nil)

	if err != nil {
		fmt.Println(err.Error())
		//fmt.Println("Domain does not exist, please double check the supplied <URL>")
		os.Exit(1)
	}

	defer tlsConn.Close()

	tlsConn.Write([]byte(req.method + " " + req.path + " HTTP/1.0\r\n"))
	tlsConn.Write([]byte("Host: " + req.host + "\r\n"))
	tlsConn.Write([]byte("\r\n"))

	data, err := ioutil.ReadAll(tlsConn)
	
	handleError(err)

	return data
}


func parseResp(data []byte) HTTPResponse {

	var status int
	headers := make(map[string]string)
	i := 0
	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	for scanner.Scan() {
		text := scanner.Text()
		if i == 0 {
			splitLn := strings.Split(text, " ")
			status, _ = strconv.Atoi(splitLn[1])
		} else {
			splitLn := strings.SplitN(text, ": ", 2)
			if len(splitLn) > 1 {
				headers[splitLn[0]] = splitLn[1]
			} else {
				break;
			}
		}
		i++
	}

	index := strings.Index(string(data), "\r\n\r\n") + 4

	return HTTPResponse{status, headers, len(data), string(data)[index:]}
}




func parseRedirect(redirect string) (URL, error) {
	r := regexp.MustCompile("^(https://|http://)")

	redirect = r.ReplaceAllString(redirect, "")


	base := redirect
	path := "/"

	index := strings.Index(redirect, "/")

	if index != -1 {
		base = redirect[:index]
		path = redirect[index:]
	}	

	return URL{base, path}, nil
}


func parseURL(url string) (URL, error) {
	url = strings.ToLower(url)

	r := regexp.MustCompile("^(https://|http://)")

	url = r.ReplaceAllString(url, "")


	base := url
	path := "/"

	index := strings.Index(url, "/")

	if index != -1 {
		path = url[index:]
		base = url[:index]
	}	

	return URL{base, path}, nil
}


func handleError(err error) {
	if err != nil {
		fmt.Printf("Error has occured: %s\n", err.Error())
		os.Exit(1)
	}
}

type URL struct {
	base string
	path string
}

type HTTPRequest struct {
	method string
	host string
	path string
}

type HTTPResponse struct {
	statusCode int
	headers map[string]string
	respSize int
	body string
}
