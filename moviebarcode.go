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

	totalFrames := getNumFrames(*srcFile)
	if numFrames > totalFrames {
		log.Fatal("Number of frames selected is greater than number of frames availible in video")
	}

	_, height := getVideoResolution(*srcFile)
	frameRate := getFrameRate(*srcFile)

	finalPixels = make([][]imageprocess.Pixel, height)
	for i := range finalPixels {
		finalPixels[i] = make([]imageprocess.Pixel, numFrames)
	}

	frames := totalFrames / numWorkers
	frameGap := totalFrames / numFrames

	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go calculateColumns(*srcFile, i*frames, (i+1)*frames-1, frameGap, frameRate, &wg)
	}
	wg.Wait()

	imageprocess.CreateImage(finalPixels)
}

func calculateColumns(filename string, start int, end int, frameGap int, frameRate float32, wg *sync.WaitGroup) {
	for i := start; i < end; i += frameGap {
		time := float32(i) / frameRate
		pixels := getFrame(filename, time)
		for j := range pixels {
			// don't know why this is neccesary
			if i/frameGap < len(finalPixels[0]) {
				finalPixels[j][i/frameGap] = imageprocess.AveragePixels(pixels[j])
			}
		}

		updateProgress()
	}
	wg.Done()
}

func getFrame(filename string, time float32) [][]imageprocess.Pixel {
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

func getFrameRate(filename string) float32 {
	cmd := exec.Command("ffprobe", "-v", "0", "-of", "csv=p=0", "-select_streams", "v:0", "-show_entries", "stream=r_frame_rate", filename)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	parts := strings.Split(strings.TrimRight(out.String(), "\r\n"), "/")
	numerator, err := strconv.Atoi(parts[0])
	denominator, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatal(err)
	}
	frameRate := float32(float64(numerator) / float64(denominator))
	return frameRate
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
