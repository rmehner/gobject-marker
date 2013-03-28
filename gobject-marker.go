package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Image struct {
	Url string `json:"url"`
}

type MarkedObject struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

var port = flag.Int("port", 8080, "the port the server listens to")
var outputFile = flag.String("outputFile", "./samples.txt", "path to the output file")
var imagePath string
var markedAbsPath string
var relPathToMarkedFromOutput string

func init() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
	}

	imagePath = filepath.Clean(flag.Arg(0))
	markedAbsPath, _ = filepath.Abs(filepath.Join(imagePath, "marked"))
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// TODO
	// * check if it is directory
	// * check if we're allowed to write in that directory
	_, err := os.Stat(imagePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}

	/*
		outputAbsPath, _ := filepath.Abs(*outputFile)
		outputAbsDir, _ := filepath.Split(outputAbsPath)

		relPathToMarkedFromOutput, relErr := filepath.Rel(outputAbsDir, markedAbsPath)
		if relErr != nil {
			fmt.Printf("Error finding relative path to marked directory: %v\n", relErr)
			os.Exit(3)
		}
	*/

	// try to create the marked images directory
	fileErr := os.Mkdir(markedAbsPath, 0666)
	if fileErr != nil && !os.IsExist(fileErr) {
		fmt.Printf("Error creating directory for marked imags: %v\n", fileErr)
		os.Exit(4)
	}

	fmt.Printf("Starting gobject marker for directory %s on http://localhost:%d\n", imagePath, *port)
	serveInterface()
}

func serveInterface() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/images/", imagesHandler)
	http.HandleFunc("/images/random", randomImageHandler)

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)

	if err != nil {
		fmt.Printf("Error serving static files: %v\n", err)
		os.Exit(5)
	}
}

func indexHandler(writer http.ResponseWriter, request *http.Request) {
	logRequest(request)

	if request.Method == "GET" && request.URL.Path == "/" {
		fmt.Fprintf(writer, "INTERFACE ...")
	} else {
		http.NotFound(writer, request)
	}
}

func imagesHandler(writer http.ResponseWriter, request *http.Request) {
	logRequest(request)

	imageName := request.URL.Path[len("/images/"):]
	pathToImage := imagePath + "/" + imageName
	method := request.Method

	if len(imageName) == 0 {
		http.NotFound(writer, request)
	} else if method == "GET" {
		http.ServeFile(writer, request, pathToImage)
	} else if method == "POST" {
		body, bodyErr := ioutil.ReadAll(request.Body)

		if bodyErr != nil {
			fmt.Fprintf(os.Stderr, "Error parsing request body: %v\n", bodyErr)
			http.Error(writer, "Unprocessable Entitiy", 422)
		}

		var markedObjects []MarkedObject
		jsonErr := json.Unmarshal(body, &markedObjects)

		if jsonErr != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", jsonErr)
			http.Error(writer, "Unprocessable Entitiy", 422)
		}

		// save to file
		// move file
	} else {
		http.Error(writer, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func randomImageHandler(writer http.ResponseWriter, request *http.Request) {
	logRequest(request)

	// ignore ALL the errors!
	// also, this is gonna be slow, because we read the dir with every request
	// will have to fix that later on
	files, _ := ioutil.ReadDir(imagePath)
	randomImage := files[randomIntWithMax(len(files))]
	imageUrl := Image{imageUrlFor(randomImage)}
	imageUrlJson, _ := json.Marshal(imageUrl)

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(writer, string(imageUrlJson))
}

func logRequest(request *http.Request) {
	log.Printf("%s %s", request.Method, request.URL.Path)
}

func randomIntWithMax(max int) int {
	// there must be a cleaner way to do this
	return int(rand.Float32() * float32(max))
}

func imageUrlFor(image os.FileInfo) string {
	return fmt.Sprintf("http://localhost:%d/images/%s", *port, image.Name())
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s [options] path-to-unmarked-images\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	os.Exit(1)
}
