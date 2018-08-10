package main

import "os"

func createFile(path string) {
	f, _ := os.Create(path)
	defer f.Close()
}
