package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/en3wton/movie-barcode-generator/imageprocess"
	"golang.org/x/image/bmp"
)

const numWorkers = 8

var finalPixels [][]imageprocess.Pixel
var numFrames int
var framesCompleted int

func main() {
	image.RegisterFormat("bmp", "bmp", bmp.Decode, bmp.DecodeConfig)

	srcFile := flag.String("filename", "", "video file to generate barcode from")
	flag.IntVar(&numFrames, "numframes", 1920, "number of frames to sample - effectively image width")
	flag.Parse()

	totalFrames := getNumFrames(*srcFile)
	if numFrames > totalFrames {
		log.Fatal("Number of frames selected is greater than number of frames availible in video")
	}

	_, height := getVideoResolution(*srcFile)

	finalPixels = make([][]imageprocess.Pixel, height)
	for i := range finalPixels {
		finalPixels[i] = make([]imageprocess.Pixel, numFrames)
	}

	frames := totalFrames / numWorkers
	frameGap := numFrames / totalFrames
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go calculateColumns(*srcFile, i*frames, (i+1)*frames, frameGap, &wg)
	}
	wg.Wait()

	imageprocess.CreateImage(finalPixels)
}

func calculateColumns(filename string, start int, end int, frameGap int, wg *sync.WaitGroup) {
	for i := start; i < end; i += frameGap {
		fmt.Println(filename, i)
		pixels := getFrame(filename, i)

		for j := range pixels {
			finalPixels[j][i] = imageprocess.AveragePixels(pixels[j])
		}

		updateProgress()
	}
	wg.Done()
}

func getFrame(filename string, index int) [][]imageprocess.Pixel {
	cmd := exec.Command("ffmpeg", "-accurate_seek", "-ss", strconv.Itoa(index), "-i",
		filename, "-frames:v", "1", "-hide_banner", "-loglevel", "0", "pipe:.bmp")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	o := bufio.NewReader(&out)
	pixels, err := imageprocess.GetPixels(o)
	if err != nil {
		fmt.Println("OOF" + strconv.Itoa(index))
		log.Fatal(err)
	}
	return pixels
}

func getNumFrames(filename string) int {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries",
		"stream=nb_frames", "-of", "default=nokey=1:noprint_wrappers=1", filename)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	string := strings.TrimRight(out.String(), "\r\n")
	n, err := strconv.Atoi(string)
	if err != nil {
		log.Fatal(err)
	}

	return n
}

func getVideoResolution(filename string) (wdith int, height int) {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries",
		"stream=width,height", "-of", "csv=s=x:p=0", filename)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	text := strings.TrimRight(out.String(), "\r\n")
	widthString := strings.Split(text, "x")[0]
	heightString := strings.Split(text, "x")[1]

	w, err := strconv.Atoi(widthString)
	h, err := strconv.Atoi(heightString)

	if err != nil {
		log.Fatal(err)
	}

	return w, h
}

func updateProgress() {
	framesCompleted++
	var percentage = float32(framesCompleted) / float32(numFrames) * 100
	fmt.Printf("\rAnalysing frames: %4.1f%%", percentage)
}

// func test(filename string, index int) {
// 	cmd := exec.Command("ffmpeg", "-accurate_seek", "-ss", strconv.Itoa(index), "-i",
// 		filename, "-frames:v", "1", "-hide_banner", "-loglevel", "0", "pipe:.bmp")
// 	var out bytes.Buffer
// 	cmd.Stdout = &out
// 	err := cmd.Run()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// o := bufio.NewReader(&out)

// outputfile, err := os.Create("test.bmp")
// if err != nil {
// 	log.Fatal(err)
// }
// defer outputfile.Close()
// io.Copy(outputfile, o)
// }
