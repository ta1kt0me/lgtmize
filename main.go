package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/disintegration/imaging"
)

const (
	Suffix   = "-lgtm"
	LGTMSize = 500
)

var (
	filter = imaging.Box
)

func main() {
	app := cli.NewApp()
	app.Name = "lgtmize"
	app.Usage = "LGTMize image"
	app.Action = func(c *cli.Context) {
		srcPath := c.Args()[0]

		src, err := imaging.Open(srcPath)
		if err != nil {
			exit(err)
		}

		resized := resize(src)
		lgtmized := drawLGTM(resized)

		if err := save(lgtmized, srcPath); err != nil {
			exit(err)
		}
	}
	app.Run(os.Args)
}

func maskPath() string {
	_, filename, _, _ := runtime.Caller(1)
	return filepath.Join(filepath.Dir(filename), "img", "mask.png")
}

func lgtmRect(img image.Image) image.Rectangle {
	size := img.Bounds().Size()

	var min, max image.Point
	switch {
	case size.X == size.Y:
		min, max = image.ZP, image.Point{LGTMSize, LGTMSize}
	case size.X > size.Y:
		min = image.Point{(size.X - LGTMSize) / 2, 0}
		max = image.Point{min.X + LGTMSize, min.Y + LGTMSize}
	case size.X < size.Y:
		min = image.Point{0, (size.Y - LGTMSize) / 2}
		max = image.Point{min.X + LGTMSize, min.Y + LGTMSize}
	}

	return image.Rectangle{min, max}
}

func resize(img image.Image) image.Image {
	size := img.Bounds().Size()

	var x, y int
	switch {
	case size.X == size.Y:
		x, y = LGTMSize, LGTMSize
	case size.X > size.Y:
		ratio := float32(size.X) / float32(size.Y)
		x, y = int(LGTMSize*ratio), LGTMSize
	case size.X < size.Y:
		ratio := float32(size.Y) / float32(size.X)
		x, y = LGTMSize, int(LGTMSize*ratio)
	}

	return imaging.Resize(img, x, y, filter)
}

func drawLGTM(img image.Image) image.Image {
	rect := img.Bounds()
	size := rect.Size()

	base := imaging.New(LGTMSize, LGTMSize, color.RGBA{255, 255, 255, 255})
	mask, err := imaging.Open(maskPath())
	if err != nil {
		exit(err)
	}

	result := imaging.New(size.X, size.Y, color.RGBA{0, 0, 0, 0})

	draw.Draw(result, rect, img, rect.Min, draw.Src)
	draw.DrawMask(result, lgtmRect(img), base, base.Bounds().Min, mask, mask.Bounds().Min, draw.Over)

	return result
}

func save(img image.Image, srcPath string) error {
	srcExt := filepath.Ext(srcPath)
	resultPath := strings.TrimRight(srcPath, srcExt) + Suffix + srcExt

	return imaging.Save(img, resultPath)
}

func exit(err error) {
	fmt.Println(err)
	os.Exit(1)
}
