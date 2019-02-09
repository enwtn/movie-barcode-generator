package main

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"sync"

	"github.com/en3wton/movie-barcode-generator/imageprocess"
	"golang.org/x/image/bmp"
)

var finalPixels [][]imageprocess.Pixel
var base = "D:\\AVGFRAME\\frames\\"
var numWorkers = 8

var numFrames int
var framesCompleted int

func main() {
	image.RegisterFormat("bmp", "bmp", bmp.Decode, bmp.DecodeConfig)
	files, _ := ioutil.ReadDir(base)
	numFrames = len(files)
	//TODO: automate get frame size
	finalPixels = make([][]imageprocess.Pixel, 816)
	for i := range finalPixels {
		finalPixels[i] = make([]imageprocess.Pixel, numFrames)
	}

	frames := numFrames / numWorkers
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go calculateColumns(files, &wg, i*frames, (i+1)*frames)
	}
	wg.Wait()

	imageprocess.CreateImage(finalPixels)
}

func calculateColumns(files []os.FileInfo, wg *sync.WaitGroup, start int, end int) {
	for i := start; i < end; i++ {
		file, err := os.Open(base + files[i].Name())
		if err != nil {
			fmt.Println("Error: File could not be opened")
			os.Exit(1)
		}
		defer file.Close()

		pixels, err := imageprocess.GetPixels(file)
		if err != nil {
			fmt.Println("Error: Image could not be decoded")
			os.Exit(1)
		}

		for j := range pixels {
			finalPixels[j][i] = imageprocess.AveragePixels(pixels[j])
		}

		updateProgress()
	}
	wg.Done()
}

func updateProgress() {
	framesCompleted++
	var percentage = float32(framesCompleted) / float32(numFrames) * 100
	fmt.Printf("\rAnalysing frames: %4.1f%%", percentage)
}
