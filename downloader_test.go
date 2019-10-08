package main

import (
	"testing"
)

func TestMissingArgument(t *testing.T) {
	args := []string{"./downloader","url","-c"}
	_,_,err := getArguments(args)
	if err == nil {
		t.Error("Expected function to return an error saying the proper usage")
	}
}

func TestNotIntArgument(t *testing.T) {
	args := []string{"./downloader","url","-c","one"}
	_,_,err := getArguments(args)
	if err == nil {
		t.Error("Expected function to return an error saying the proper usage")
	}
}	

func TestGoodArgument(t *testing.T) {
	args := []string{"./downloader","url","-c","1"}
	s,i,err := getArguments(args)
	if s != "url" || i != 1 || err != nil {
		t.Errorf("Arguments were not properly parsed; wanted url: url nThreads: 1 err: nil, got url: %v nThreads: %v err: %v",s,i,err)
	}
}	
