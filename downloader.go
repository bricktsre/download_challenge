package main

import ( 
	"net/http"
	"io"
	"strconv"
	"os"
	"errors"
	"strings"
)

// Holds information necessary for a individual goroutine to download a chunk of a file
type downloader struct{
	start int
	end int
	url string
	outfile	*os.File
	err chan error
	done chan error
}

/*
  Downloads a file from a given url concurrently using one or more threads
*/
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
	fileLength, err := strconv.Atoi(maps["Content-Length"][0])   // get length of file to be downloaded
	if err != nil {
		panic(err)
	}
	
	filepath := url[strings.LastIndex(url, "/")+1: len(url)]  // parse name to use for downloaded file
	if len(filepath) < 1 {
		filepath = "file.txt"
	}
	
	outFile, err := os.Create(filepath + ".tmp")   // create a temporary file to write downloaded chunks into
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	chunkSize := fileLength / numThreads   // size of chunk each thread will download
	
	downloadarr := make([]downloader, numThreads)	// array holding each goroutine
	errchan := make(chan error)
	donechan := make(chan error)
	for i := 0; i < numThreads; i++ {
		downloadarr[i] = downloader{ 		// create a download worker
			start: i*chunkSize,
			end: (i+1)*chunkSize,
			url: url, 
			outfile: outFile, 
			err: errchan,
			done: donechan,
		}
		 if i == numThreads-1 {
			downloadarr[i].end = fileLength	  // special case where last chunk may need to be slightly larger to download the whole file
		}
		go downloadarr[i].download()		// start the download goroutine
	}
	
	count := 0
	errorloop:for {		// loop waiting for all goroutines to finish
		select {
			case err = <-errchan:
				panic(err)	// panic if a goroutine sends an error
			case <- donechan:
				count++		// a goroutine has signalled it is done so increment count of finished goroutines
				if count == numThreads {
					break errorloop	  // If all goroutines are done break out of loop
				}
		}
	}

	err = os.Rename(filepath + ".tmp", filepath)	// change file extension from temporary 
	if err != nil {
		panic(err)
	}
}

/*
   Grabs aruments from the command line
   string: url of file
   int: number of threads to use
   error: error if one occured, nil otherwise
*/
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

/*
   Downloads a chunk of a file
   The bytes to download are specified by d.start-d.end
   Calls writeOut() to write the downloaded chunk to a file
   Is run from a goroutine
*/
func (d downloader) download() {	
	client := &http.Client {}
	req, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		d.err <- err		// signall main process this goroutine has encountered an error
		return
	}
	range_header := "bytes=" + strconv.Itoa(d.start) + "-" + strconv.Itoa(d.end-1)  // bytes to grab
	req.Header.Add("Range", range_header)   // adding byte request to http request1
	
	resp, err := client.Do(req)
	if err != nil {
		d.err <- err
		return
	}
	defer resp.Body.Close()
	
	err = d.writeOut(resp.Body)		// function writing to file
	if err != nil {
		d.err <- err
		return
	}
	d.done <- nil		// signal main process this goroutine is done
}

/* 
   Writes out the http Response body to the file
   body is an io.ReadCloser from the body of an http response
   Returns an error if an error occured or nil if the write out is 
   successful
*/
func (d downloader) writeOut(body io.ReadCloser) error {
	buf := make([]byte, 4*1024)		// temporary buffer to read into and write from
	for {
		br, err := body.Read(buf)	// read up to buf bytes
		if br > 0 {
			bw, err := d.outfile.WriteAt(buf[0:br], int64(d.start)) // write out to the file starting at a specific offest
			if err != nil {
				return err
			}
			if br != bw {
				errors.New("Not all bytes read were written")
			}

			d.start = bw + d.start		// increment the file write offset
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
	}
	return nil
}
