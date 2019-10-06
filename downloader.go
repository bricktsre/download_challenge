package main

import ( 
	"fmt"
	"net/http"
	"io"
	"strconv"
	"sync"
	"os"
	"errors"
	"strings"
)

type downloader struct{
	start int
	end int
	url string
	outfile	*os.File
	err chan error
	done chan error
	fileLock sync.Mutex
}

type OffsetWriter struct{
	file *os.File
	offset int64
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
	
	outFile, err := os.Create(filepath + ".tmp")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	chunkSize := fileLength / numThreads
	//lastChunk := fileLength % numThreads
	
	downloadarr := make([]downloader, numThreads)	
	errchan := make(chan error)
	donechan := make(chan error)
	for i := 0; i < numThreads; i++ {
		downloadarr[i] = downloader{
			start: i*chunkSize,
			end: (i+1)*chunkSize,
			url: url, 
			outfile: outFile, 
			err: errchan,
			done: donechan,
			fileLock: sync.Mutex{} }
		if i == numThreads-1 {
			downloadarr[i].end = fileLength
		}
		go downloadarr[i].download()
	}
	
	count := 0
	errorloop:for {
		select {
			case err = <-errchan:
				panic(err)
			case <- donechan:
				count++
				if count == numThreads {
					break errorloop
				}
		}
	}

	err = os.Rename(filepath + ".tmp", filepath)
	if err != nil {
		panic(err)
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

func (dst *OffsetWriter) Write(b []byte) (n int, err error) {
	if dst.offset == 13292 {
		fmt.Printf("%x %x\n",b[0] ,b[1])
	}
	n, err = dst.file.WriteAt(b, dst.offset)
	dst.offset += int64(n)
	fmt.Println(dst.offset-int64(n), len(b), n)
	return	
}

func (d downloader) download() {	
	client := &http.Client {}
	req, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		d.err <- err
		return
	}
	range_header := "bytes=" + strconv.Itoa(d.start) + "-" + strconv.Itoa(d.end-1)
	req.Header.Add("Range", range_header)
	fmt.Println(range_header)	
	resp, err := client.Do(req)
	if err != nil {
		d.err <- err
		return
	}
	defer resp.Body.Close()
	
	err = d.writeOut(resp.Body)
	/*writer := &OffsetWriter{file: d.outfile, offset: int64(d.start)}
	d.fileLock.Lock()
	fmt.Println("going to write")
	_, err = io.Copy(d.outfile, io.TeeReader(resp.Body, writer))
	d.fileLock.Unlock()
	*/if err != nil {
		d.err <- err
		return
	}
	d.done <- nil	
}

func (d downloader) writeOut(body io.ReadCloser) error {
	buf := make([]byte, 4*1024)
	for {
		br, err := body.Read(buf)
		if br > 0 {
			bw, err := d.outfile.WriteAt(buf[0:br], int64(d.start))
			if err != nil {
				return err
			}
			if br != bw {
				errors.New("Not all bytes read were written")
			}

			d.start = bw + d.start
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
