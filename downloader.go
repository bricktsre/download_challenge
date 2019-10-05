package main

import ( 
	"fmt"
	//"net/http"
	//"io/ioutil"
	"strconv"
	//"sync"
	"os"
)

func main() {
	if len(os.Args) < 4 || len(os.Args) > 5{
		fmt.Println("Proper usage: ./downloader <url> -c numberOfThreads")
		return	
	}
	url := os.Args[1]
	
	if os.Args[2] != "-c" {
		fmt.Println("Proper usage: ./downloader <url> -c numberOfThreads")
		return
	}
	numThreads, err := strconv.Atoi(os.Args[3])
	if err != nil { 
		panic(err)
	}
	fmt.Println(url, numThreads)
}
