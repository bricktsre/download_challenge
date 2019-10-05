package main

import ( 
	//"fmt"
	"net/http"
	"io"
	"strconv"
	//"sync"
	"os"
	"errors"
	"strings"
)

type downloader struct{
	start int
	end int
	url string
	filepath string
	err chan error
}

func main() {
	url, numThreads, err := getArguments()
	if err != nil {
		panic(err)
	}

	res, err := http.Head(url)
	if err != nil {
		panic(err)
	}
	maps := res.Header
	fileLength, err := strconv.Atoi(maps["Content-Length"][0])
	if err != nil {
		panic(err)
	}
	
	filepath := url[strings.LastIndex(url, "/")+1: len(url)]
	if len(filepath) < 1 {
		filepath = "file.txt"
	}

	chunkSize := fileLength / numThreads
	lastChunk := fileLength % numThreads
	lastChunk++	

	testd := downloader{0,chunkSize, url, filepath, make(chan error, 1)}
	go download(testd)
	for e := range testd.err {
		panic(e)
	}
}

func getArguments() (string, int, error) {
	if len(os.Args) < 4 || len(os.Args) > 5{
		return "", 0, errors.New("Proper usage: ./downloader <url> -c numberOfThreads")
			
	}
	url := os.Args[1]
	
	if os.Args[2] != "-c" {
		return "", 0, errors.New("Proper usage: ./downloader <url> -c numberOfThreads")
	}
	numThreads, err := strconv.Atoi(os.Args[3])
	if err != nil { 
		return "", 0, err
	}
	return url, numThreads, nil
}

func download(d downloader) {
	defer close(d.err)
		
	out, err := os.Create(d.filepath + ".tmp")
	if err != nil {
		d.err <- err
		return
	}
	defer out.Close()

	client := &http.Client {}
	req, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		d.err <- err
		return
	}
	req.Header.Add("Range", "bytes=" + strconv.Itoa(d.start) + "-" + strconv.Itoa(d.end-1))
	
	resp, err := client.Do(req)
	if err != nil {
		d.err <- err
		return
	}
	defer resp.Body.Close()
	
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		d.err <- err
		return
	}
	
	err = os.Rename(d.filepath + ".tmp", d.filepath)
	if err != nil {
		d.err <- err
		return
	}
}
