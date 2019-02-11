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
Jumper (2008)
![Jumper](https://raw.githubusercontent.com/en3wton/movie-barcode-generator/master/example-images/jumper.png "Jumper (2008)")

Iron Man (2008)
![Iron Man](https://raw.githubusercontent.com/en3wton/movie-barcode-generator/master/example-images/ironman.png "Iron Man (2008)")

Ace Ventura: Pet Detective (1994)
![Ace Ventura](https://raw.githubusercontent.com/en3wton/movie-barcode-generator/master/example-images/aceventura.png "Ace Ventura: Pet Detective (1994)")
