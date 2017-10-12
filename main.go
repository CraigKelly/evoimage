package main

import (
	"errors"
	"flag"
	"log"
	"os"
)

func pcheck(err error) {
	if err != nil {
		log.Panicf("Fatal Error: %v\n", err)
	}
}

/////////////////////////////////////////////////////////////////////////////
// Entry point

func main() {
	flags := flag.NewFlagSet("evoimage", flag.ExitOnError)
	mutationRate := flags.Float64("mutationRate", 0.01, "Mutation rate to use")
	crossOverRate := flags.Float64("crossoverRate", 0.7, "Crossover rate to use")
	popSize := flags.Int("popSize", 300, "Population size in a generation")
	image := flags.String("image", "", "File name of target image")

	pcheck(flags.Parse(os.Args[1:]))

	if *mutationRate <= 0.0 || *mutationRate >= 1.0 {
		pcheck(errors.New("Invalid mutation rate - must be between 0 and 1"))
	}
	if *crossOverRate <= 0.0 || *crossOverRate >= 1.0 {
		pcheck(errors.New("Invalid crossover rate - must be between 0 and 1"))
	}
	if *popSize < 10 {
		pcheck(errors.New("Invalid population size - must be at least 10"))
	}
	if image == nil || len(*image) < 1 {
		pcheck(errors.New("Image filename is required"))
	}
	if _, err := os.Stat(*image); err != nil {
		pcheck(err)
	}

	log.Printf("Mutation:%f, Crossover:%f, Population:%d, Target:%s\n", *mutationRate, *crossOverRate, *popSize, *image)

	os.Exit(0)
}
