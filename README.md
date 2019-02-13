# Movie Barcode Editor
![Jumper](https://raw.githubusercontent.com/en3wton/movie-barcode-generator/master/example-images/jumper-header.png "Jumper 2008")

A program to generate "Movie Barcodes".  
Idea stolen from http://moviebarcode.tumblr.com/ and various others

**Requirements**
* ffmpeg
* ffprobe

## Usage
```movie-barcode-generator (.exe) -filename <filename> -numframes <number of frames to sample>```

Note that the image width will be equal to the number of frames sampled.  
The image height is equal to the height of the video

**Currently only works on mp4 files**

## Examples
The Florida Project (2017)
![The Florida Project](https://raw.githubusercontent.com/en3wton/movie-barcode-generator/master/example-images/floridaproject.png "The Florida Project (2017)")

Jumper (2008)
![Jumper](https://raw.githubusercontent.com/en3wton/movie-barcode-generator/master/example-images/jumper.png "Jumper (2008)")

Iron Man (2008)
![Iron Man](https://raw.githubusercontent.com/en3wton/movie-barcode-generator/master/example-images/ironman.png "Iron Man (2008)")