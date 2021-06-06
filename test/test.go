package main

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"log"
	"os"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func HLine(img draw.Image, x1, y, x2 int, col color.Color) {
	for ; x1 <= x2; x1++ {
		img.Set(x1, y, col)
	}
}

func VLine(img draw.Image, x, y1, y2 int, col color.Color) {
	for ; y1 <= y2; y1++ {
		img.Set(x, y1, col)
	}
}

func Rect(img draw.Image, x1, y1, w, h int, col color.Color) {
	HLine(img, x1, y1, x1+w, col)
	HLine(img, x1, y1+h, x1+w, col)
	VLine(img, x1, y1, y1+h, col)
	VLine(img, x1+w, y1, y1+h, col)
}

func drawBoxes(fname string, x float32, y float32, w float32, h float32) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}

	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, b, img, b.Min, draw.Src)
	col := color.RGBA{0, 255, 0, 255}

	// HLine(dst, 10, 200, 200, col)

	log.Println(int(float32(b.Dx())*x), int(float32(b.Dy())*y), int(float32(b.Dx())*w), int(float32(b.Dy())*h))
	Rect(dst, int(float32(b.Dx())*x), int(float32(b.Dy())*y), int(float32(b.Dx())*w), int(float32(b.Dy())*h), col)

	// f, err = os.Create(fname)
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()

	// opt := jpeg.Options{
	// 	Quality: 100,
	// }
	// err = jpeg.Encode(f, dst, &opt)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func main() {
	// 	log.Println(">")
	// 	start := time.Now()
	// 	err := drawBoxes("./test/in.jpeg", 0.099000, 0.366000, 0.066000, 0.061000)
	// 	elapsed := time.Since(start)
	// 	log.Printf("took %s", elapsed)
	// 	err = drawBoxes("./test/in.jpeg", 0.496000, 0.253000, 0.057000, 0.075000)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	err := ffmpeg.Input("file.mpeg").Output("file.mp4", ffmpeg.KwArgs{
		"format": "mp4",
		"vcodec": "libx264", "preset": "ultrafast", "acodec": "none", "movflags": "+faststart",
	}).OverWriteOutput().Run()
	log.Println(err)
}
