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

// TODO: make sure adaptive stuff is properly documented
// TODO: update docs about fitness scaled by max fitness
// TODO: update docs about guassian mutation
// TODO: use triangles instead of rectangle
// TODO: allow size limiting of rectangles AND report size across solutions
// TODO: need table with triangle count and size with outputs
// TODO: consider a goroutine for logging once we add triangle size/count to log

//////////////////////////////////////////////////////////////////////////
// Helpers

// colorDist return a positive measure of distance between two colors
// currently this is Euclidean distance ignoring Alpha
func colorDist(c1 color.NRGBA, c2 color.NRGBA) float64 {
	rd := math.Pow(float64(c1.R)-float64(c2.R), 2.0)
	gd := math.Pow(float64(c1.G)-float64(c2.G), 2.0)
	bd := math.Pow(float64(c1.B)-float64(c2.B), 2.0)

	return math.Sqrt(rd + gd + bd)
}

//////////////////////////////////////////////////////////////////////////
// Our target image

// ImageTarget is the image we are actually trying to reproduce
type ImageTarget struct {
	fileName   string
	imageData  *image.NRGBA
	imageMode  *color.NRGBA
	imageMean  *color.NRGBA
	maxFitness float64
}

// NewImageTarget creates a new ImageTarget instance from the JPEG file
func NewImageTarget(fileName string) (*ImageTarget, error) {
	fimg, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer fimg.Close()

	simg, err := jpeg.Decode(fimg)
	if err != nil {
		return nil, err
	}

	// Make sure that the image is actually in NRGBA format
	b := simg.Bounds()
	img := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(img, img.Bounds(), simg, b.Min, draw.Src)

	// Calculate max fitness
	b = img.Bounds()
	yrng := (b.Max.Y - b.Min.Y) + 1
	xrng := (b.Max.X - b.Min.X) + 1
	pixCount := float64(xrng * yrng)
	oneMax := math.Sqrt(255.0 * 255.0 * 3.0) // 255 squared times 3 for RGB
	maxFit := pixCount * oneMax

	log.Printf("%s %v %v (mf=%f)\n", fileName, img.ColorModel(), img.Bounds(), maxFit)

	return &ImageTarget{
		fileName:   fileName,
		imageData:  img,
		maxFitness: maxFit,
	}, nil
}

func (it *ImageTarget) calcStats() {
	counts := make(map[color.NRGBA]uint)
	bnd := it.imageData.Bounds()
	pixCount := 0
	r := float64(0.0)
	g := float64(0.0)
	b := float64(0.0)
	a := float64(0.0)

	var clr color.NRGBA
	for y := bnd.Min.Y; y < bnd.Max.Y; y++ {
		for x := bnd.Min.X; x < bnd.Max.X; x++ {
			clr = it.imageData.NRGBAAt(x, y)
			counts[clr]++
			pixCount++
			r += float64(clr.R)
			g += float64(clr.G)
			b += float64(clr.B)
			a += float64(clr.A)
		}
	}

	var modeClr color.NRGBA
	var modeCount uint
	for clr, count := range counts {
		if count > modeCount {
			modeClr = clr
			modeCount = count
		}
	}

	pc := float64(pixCount)
	meanClr := color.NRGBA{
		R: uint8(r / pc),
		G: uint8(g / pc),
		B: uint8(b / pc),
		A: uint8(a / pc),
	}

	log.Printf("Colors => mean=%v, mode=%v with %d occurs\n", meanClr, modeClr, modeCount)
	it.imageMode = &modeClr
	it.imageMean = &meanClr
}

// ImageMode returns the most common color in the image (use as a background color)
func (it *ImageTarget) ImageMode() color.NRGBA {
	if it.imageMode == nil {
		it.calcStats()
	}

	return *it.imageMode
}

// ImageMean return the average color (by averaging each color channel)
func (it *ImageTarget) ImageMean() color.NRGBA {
	if it.imageMean == nil {
		it.calcStats()
	}

	return *it.imageMean
}

//////////////////////////////////////////////////////////////////////////
// Genes - single encoded feature

// Gene represents single item in a genome
type Gene struct {
	destBounds *image.Rectangle
	destColor  *color.NRGBA
}

// NewGene creates a random gene instance
func NewGene(src *ImageTarget) *Gene {
	b := src.imageData.Bounds()
	yrng := (b.Max.Y - b.Min.Y) + 1
	xrng := (b.Max.X - b.Min.X) + 1

	// Copy the canonical rectangle for our pointer below
	var rct image.Rectangle
	rct = image.Rect(
		rand.Intn(xrng)+b.Min.X, rand.Intn(yrng)+b.Min.Y,
		rand.Intn(xrng)+b.Min.X, rand.Intn(yrng)+b.Min.Y,
	)

	var clr color.NRGBA = color.NRGBA{
		R: uint8(rand.Intn(256)),
		G: uint8(rand.Intn(256)),
		B: uint8(rand.Intn(256)),
		A: uint8(rand.Intn(8)),
	}

	return &Gene{
		destBounds: &rct,
		destColor:  &clr,
	}
}

// Copy returns a pointer to a proper deep copy of a Gene
func (g *Gene) Copy() *Gene {
	newg := Gene{
		destBounds: new(image.Rectangle),
		destColor:  new(color.NRGBA),
	}
	*newg.destBounds = *g.destBounds
	*newg.destColor = *g.destColor
	return &newg
}

//////////////////////////////////////////////////////////////////////////
// Our candidate image - aka an individual genome, made up of Gene's

// Individual is a single candidate individual in a population
type Individual struct {
	target    *ImageTarget
	fitness   float64
	imageData image.Image
	needImage bool
	genes     [64]*Gene
}

// NewIndividual creates a random individual
func NewIndividual(src *ImageTarget) *Individual {
	// For now we have a fixed genome
	ind := Individual{
		target:    src,
		fitness:   -1.0,
		needImage: true,
	}
	return &ind
}

// RandInit initializes the individual to a random state
func (ind *Individual) RandInit() {
	for i := 0; i < len(ind.genes); i++ {
		ind.genes[i] = NewGene(ind.target)
	}
}

// Fitness calculates the individual's fitness score (to be minimized) using lazy and cached evaluation
func (ind *Individual) Fitness() float64 {
	if !ind.needImage {
		return ind.fitness
	}

	// TODO: when we switch to rectangles, we'll be drawing to RGBA (although we
	// can still use the NRGBA for each polygon since the stroke color is set via
	// the color interface). We'll need to do a final image copy to NRGBA to make
	// sure that our pixel colors are correct for the fitness loop below

	// init image: color entire rectange from src.ImageMode
	img := image.NewNRGBA(ind.target.imageData.Bounds())
	draw.Draw(img, img.Bounds(), &image.Uniform{ind.target.ImageMode()}, image.ZP, draw.Src)

	// Now we need to draw all the rectangles in our genome
	for _, gene := range ind.genes {
		draw.Draw(img, *gene.destBounds, &image.Uniform{gene.destColor}, image.ZP, draw.Over)
	}

	// calculate fitness - the sum of the color distance pixel by pixel
	fitness := float64(0.0)

	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c1 := img.NRGBAAt(x, y)
			c2 := ind.target.imageData.NRGBAAt(x, y)
			fitness += colorDist(c1, c2)
		}
	}

	// Scale by the maxmimum error
	fitness = (fitness / ind.target.maxFitness) * 100.0

	// all done - store our results and return the fitness
	ind.fitness = fitness
	ind.imageData = img
	ind.needImage = false
	return ind.fitness
}

// Save the individual as a JPEG using the given file name
func (ind *Individual) Save(fileName string) error {
	fimg, ferr := os.Create(fileName)
	if ferr != nil {
		return ferr
	}
	defer fimg.Close()

	opts := &jpeg.Options{
		Quality: 99,
	}

	ierr := jpeg.Encode(fimg, ind.imageData, opts)
	if ierr != nil {
		return ierr
	}

	//log.Printf("Wrote file %s\n", fileName)

	return nil
}

//////////////////////////////////////////////////////////////////////////
// Sort order for individual

// Population is a collection type that provides sorting and some helpers
type Population []*Individual

func (a Population) Len() int           { return len(a) }
func (a Population) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Population) Less(i, j int) bool { return a[i].Fitness() < a[j].Fitness() }

// MeanFitness calculates arithmetic mean of the population
func (a Population) MeanFitness() float64 {
	tot := float64(0.0)
	for _, i := range a {
		tot += i.Fitness()
	}
	return tot / float64(len(a))
}
