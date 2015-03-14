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

		lgtmized, err := drawLGTM(resized)
		if err != nil {
			exit(err)
		}

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
		min, max = image.ZP, image.Pt(LGTMSize, LGTMSize)
	case size.X > size.Y:
		min = image.Pt((size.X-LGTMSize)/2, 0)
		max = image.Pt(min.X+LGTMSize, min.Y+LGTMSize)
	case size.X < size.Y:
		min = image.Pt(0, (size.Y-LGTMSize)/2)
		max = image.Pt(min.X+LGTMSize, min.Y+LGTMSize)
	}

	return image.Rect(min.X, min.Y, max.X, max.Y)
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

func drawLGTM(img image.Image) (image.Image, error) {
	rect := img.Bounds()
	size := rect.Size()
	result := imaging.New(size.X, size.Y, color.RGBA{0, 0, 0, 0})
	draw.Draw(result, rect, img, rect.Min, draw.Src)

	lgtm := imaging.New(LGTMSize, LGTMSize, color.RGBA{255, 255, 255, 255})
	mask, err := imaging.Open(maskPath())
	if err != nil {
		return nil, err
	}
	draw.DrawMask(result, lgtmRect(img), lgtm, lgtm.Bounds().Min, mask, mask.Bounds().Min, draw.Over)

	return result, nil
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
