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
var outputAbsPath string

func init() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
	}

	imagePath = filepath.Clean(flag.Arg(0))
	markedAbsPath, _ = filepath.Abs(filepath.Join(imagePath, "marked"))
	rand.Seed(time.Now().UnixNano())

	outputAbsPath, _ = filepath.Abs(*outputFile)
	outputAbsDir, _ := filepath.Split(outputAbsPath)
	relPathToMarkedFromOutput, _ = filepath.Rel(outputAbsDir, markedAbsPath)
}

func main() {
	// @TODO
	// * check if it is directory
	// * check if we're allowed to write in that directory
	_, err := os.Stat(imagePath)
	if err != nil {
		logErrorAndExit(err, 2)
	}

	// try to create the marked images directory
	fileErr := os.Mkdir(markedAbsPath, 0666)
	if fileErr != nil && !os.IsExist(fileErr) {
		logErrorAndExit(fileErr, 3)
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
		logErrorAndExit(err, 4)
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
	pathToImage := filepath.Join(imagePath, imageName)
	method := request.Method

	if len(imageName) == 0 {
		http.NotFound(writer, request)
	} else if method == "GET" {
		http.ServeFile(writer, request, pathToImage)
	} else if method == "POST" {
		body, err := ioutil.ReadAll(request.Body)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing request body: %v\n", err)
			http.Error(writer, "Unprocessable Entitiy", 422)
		}

		var markedObjects []MarkedObject
		err = json.Unmarshal(body, &markedObjects)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			http.Error(writer, "Unprocessable Entitiy", 422)
		}

		markedImagePath := filepath.Join(relPathToMarkedFromOutput, imageName)
		outputString := fmt.Sprintf("%s %d", markedImagePath, len(markedObjects))

		for _, el := range markedObjects {
			outputString += fmt.Sprintf(" %d %d %d %d", el.X, el.Y, el.Width, el.Height)
		}

		outputString += "\n"
		_, err = appendToOutputfile(outputString)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			http.Error(writer, "Error writing to output file", http.StatusInternalServerError)
		} else {
			http.Error(writer, "Marked objects added", http.StatusCreated)

			err = os.Rename(pathToImage, markedImagePath)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error moving file: %v\n", err)
			}
		}
	} else {
		http.Error(writer, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func randomImageHandler(writer http.ResponseWriter, request *http.Request) {
	logRequest(request)

	// @TODO
	// * Do not ignore errors
	// * better way to get random file in directory, current way will be slow
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

func logErrorAndExit(err error, exitCode int) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(exitCode)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s [options] path-to-unmarked-images\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func appendToOutputfile(outputString string) (ret int, err error) {
	// open up output file
	outputFileHandle, err := os.OpenFile(outputAbsPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening output file %s: %v\n", outputAbsPath, err)
	}

	// @TODO remember fileHandle to prevent reopening this file again & again
	defer outputFileHandle.Close()

	return outputFileHandle.WriteString(outputString)
}
