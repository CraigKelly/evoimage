package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"math"
	"math/rand"
	"os"
)

//////////////////////////////////////////////////////////////////////////
// Helpers

// colorDist return a positive measure of distance between two colors
// currently this is Euclidean distance
func colorDist(c1 color.Color, c2 color.Color) float64 {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	rd := math.Pow(float64(r1)-float64(r2), 2.0)
	gd := math.Pow(float64(g1)-float64(g2), 2.0)
	bd := math.Pow(float64(b1)-float64(b2), 2.0)
	ad := math.Pow(float64(a1)-float64(a2), 2.0)

	return math.Sqrt(rd + gd + bd + ad)
}

//////////////////////////////////////////////////////////////////////////
// Our target image

type ImageTarget struct {
	fileName  string
	imageData image.Image
	imageMode *color.Color
}

func NewImageTarget(fileName string) (ImageTarget, error) {
	fimg, err := os.Open(fileName)
	if err != nil {
		return ImageTarget{}, err
	}
	defer fimg.Close()

	img, err := jpeg.Decode(fimg)
	if err != nil {
		return ImageTarget{}, err
	}

	log.Printf("%s %v %v\n", fileName, img.ColorModel(), img.Bounds())

	return ImageTarget{
		fileName:  fileName,
		imageData: img,
	}, nil
}

func (it ImageTarget) ImageMode() color.Color {
	if it.imageMode != nil {
		return *it.imageMode
	}

	counts := make(map[color.Color]uint)
	b := it.imageData.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			clr := it.imageData.At(x, y)
			counts[clr] += 1
		}
	}

	var modeClr color.Color
	var modeCount uint
	for clr, count := range counts {
		if count > modeCount {
			modeClr = clr
			modeCount = count
		}
	}

	log.Printf("Most frequent color is %v with %d occurs\n", modeClr, modeCount)
	it.imageMode = &modeClr
	return modeClr
}

//////////////////////////////////////////////////////////////////////////
// Genes - single encoded feature

type Gene struct {
	destBounds image.Rectangle
	destColor  color.Color
}

func NewGene(src ImageTarget) Gene {
	b := src.imageData.Bounds()
	yrng := (b.Max.Y - b.Min.Y) + 1
	xrng := (b.Max.X - b.Min.X) + 1
	pt1 := image.Pt(rand.Intn(xrng)+b.Min.X, rand.Intn(yrng)+b.Min.Y)
	pt2 := image.Pt(rand.Intn(xrng)+b.Min.X, rand.Intn(yrng)+b.Min.Y)

	clr := color.RGBA{
		uint8(rand.Intn(256)),
		uint8(rand.Intn(256)),
		uint8(rand.Intn(256)),
		uint8(rand.Intn(256)),
	}

	return Gene{
		destBounds: image.Rectangle{pt1, pt2}.Canon(),
		destColor:  clr,
	}
}

//////////////////////////////////////////////////////////////////////////
// Our candidate image - aka an individual genome, made up of Gene's

type Individual struct {
	fitness   float64
	imageData image.Image
	needImage bool
	genes     [200]Gene
}

func NewIndividual(src ImageTarget) *Individual {
	// For now we have a fixed genome
	ind := Individual{
		fitness:   -1.0,
		needImage: true,
	}
	for i := 0; i < len(ind.genes); i++ {
		ind.genes[i] = NewGene(src)
	}
	return &ind
}

func (i *Individual) Fitness(src ImageTarget) float64 {
	if !i.needImage {
		return i.fitness
	}

	// init image: color entire rectange from src.ImageMode
	img := image.NewRGBA(src.imageData.Bounds())
	draw.Draw(img, img.Bounds(), &image.Uniform{src.ImageMode()}, image.ZP, draw.Src)

	// Now we need to draw all the rectangles in our genome
	for _, gene := range i.genes {
		draw.Draw(img, gene.destBounds, &image.Uniform{gene.destColor}, image.ZP, draw.Src)
	}

	// calculate fitness - the sum of the color distance pixel by pixel
	fitness := float64(0.0)

	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c1 := img.At(x, y)
			c2 := src.imageData.At(x, y)
			fitness += colorDist(c1, c2)
		}
	}

	// all done - store our results and return the fitness
	i.fitness = fitness
	i.imageData = img
	i.needImage = false
	return i.fitness
}

func (i *Individual) Save(fileName string) error {
	fimg, ferr := os.Create(fileName)
	if ferr != nil {
		return ferr
	}
	defer fimg.Close()

	opts := &jpeg.Options{
		Quality: 99,
	}

	ierr := jpeg.Encode(fimg, i.imageData, opts)
	if ierr != nil {
		return ierr
	}

	return nil
}
