package main

import (
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

	// TODO Check number of frames does not exceed total number of frames

	_, height := getVideoResolution(*srcFile)

	finalPixels = make([][]imageprocess.Pixel, height)
	for i := range finalPixels {
		finalPixels[i] = make([]imageprocess.Pixel, numFrames)
	}

	length := getVideoLength(*srcFile)
	interval := length / float64(numFrames)
	period := length / float64(numWorkers)

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go calculateColumns(*srcFile, float64(i)*period, float64(i+1)*period, interval, &wg)
	}
	wg.Wait()

	imageprocess.CreateImage(finalPixels)
}

func calculateColumns(filename string, start float64, end float64, interval float64, wg *sync.WaitGroup) {
	for i := start; i < end; i += interval {
		pixels := getFrame(filename, i)
		for j := range pixels {
			finalPixels[j][int(i/interval)] = imageprocess.AveragePixels(pixels[j])
		}

		updateProgress()
	}
	wg.Done()
}

func getFrame(filename string, time float64) [][]imageprocess.Pixel {
	t := fmt.Sprintf("%f", time)
	cmd := exec.Command("ffmpeg", "-accurate_seek", "-ss", t, "-i",
		filename, "-frames:v", "1", "-hide_banner", "-loglevel", "0", "pipe:.bmp")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	pixels, err := imageprocess.GetPixels(&out)
	if err != nil {
		fmt.Println("Error at:" + t)
		log.Fatal(err)
	}
	return pixels
}

func getVideoLength(filename string) float64 {
	cmd := exec.Command("ffprobe", "-i", filename, "-show_entries", "format=duration", "-v", "quiet", "-of", "csv=p=0")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	length, err := strconv.ParseFloat(strings.TrimRight(out.String(), "\r\n"), 64)
	if err != nil {
		log.Fatal(err)
	}

	return length
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
