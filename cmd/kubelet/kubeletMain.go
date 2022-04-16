package main

import "os"

func main() {
	os.MkdirAll("./tmp", os.ModePerm)
}
