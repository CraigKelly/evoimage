package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"runtime"
	"sort"
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
	popSize := flags.Int("popSize", 100, "Population size in a generation")
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

	log.Printf("Loading image %s\n", *image)
	target, err := NewImageTarget(*image)
	pcheck(err)
	target.ImageMode()

	log.Printf("Creating init pop of %d\n", *popSize)
	population := Population(make([]*Individual, 0, *popSize))
	for i := 0; i < *popSize; i++ {
		population = append(population, NewIndividual(target))
	}

	cores := runtime.NumCPU()
	if cores < 2 {
		cores = 2
	}
	log.Printf("Working with %d cores\n", cores)

	log.Printf("Evaluating...\n")
	work := make(chan int, 8)
	for c := 0; c < cores; c++ {
		go func() {
			for idx := range work {
				population[idx].Fitness()
			}
		}()
	}
	for i := range population {
		work <- i
	}
	close(work)

	log.Printf("Sorting...\n")
	sort.Sort(population)
	log.Printf("Best  Individual: fit %.2f => latest.jpg\n", population[0].Fitness())
	log.Printf("Worst Individual: fit %.2f\n", population[len(population)-1].Fitness())
	log.Printf("             Mean fit %.2f\n", population.MeanFitness())

	population[0].Save("latest.jpg")

	// TODO: image init and comparison

	os.Exit(0)
}
