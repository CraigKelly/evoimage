package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"
)

// helper for checking errors
func pcheck(err error) {
	if err != nil {
		log.Panicf("Fatal Error: %v\n", err)
	}
}

// multi-core evaluation
func evalPop(pop Population, cores int) {
	wait := sync.WaitGroup{}
	wait.Add(cores)

	work := make(chan int, 8)

	for c := 0; c < cores; c++ {
		go func() {
			defer wait.Done()
			for idx := range work {
				pop[idx].Fitness()
			}
		}()
	}

	for i := range pop {
		work <- i
	}
	close(work)

	wait.Wait()
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

	rand.Seed(time.Now().UnixNano())

	log.Printf("Loading image %s\n", *image)
	target, err := NewImageTarget(*image)
	pcheck(err)
	target.ImageMode()

	logFileName := fmt.Sprintf("logs/%s-log.csv", *image)
	log.Printf("Opening log file %s\n", logFileName)
	logf, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	pcheck(err)
	dataLog := csv.NewWriter(logf)
	defer dataLog.Flush()
	defer logf.Close()
	// Always write a title line - that way we can detect restarts
	pcheck(dataLog.Write([]string{"Gen", "Best", "Worst", "Avg", "Timestamp"}))

	log.Printf("Creating init pop of %d\n", *popSize)
	population := Population(make([]*Individual, 0, *popSize))
	for i := 0; i < *popSize; i++ {
		ind := NewIndividual(target)
		ind.RandInit()
		population = append(population, ind)
	}

	cores := runtime.NumCPU()
	if cores < 2 {
		cores = 2
	}
	log.Printf("Working with %d cores\n", cores)

	for generation := 0; generation < 100000; generation++ {
		log.Printf("Generation %d\n", generation)

		// Image creation and evaluation across all cores
		evalPop(population, cores)

		log.Printf("Sorting...\n")
		sort.Sort(population)
		best := population[0].Fitness()
		worst := population[len(population)-1].Fitness()
		avg := population.MeanFitness()
		log.Printf("Best  Individual: fit %.2f (L %.4f) => latest.jpg\n", best, math.Log(best))
		log.Printf("Worst Individual: fit %.2f (L %.4f)\n", worst, math.Log(worst))
		log.Printf("             Mean fit %.2f (L %.4f)\n", avg, math.Log(avg))

		pcheck(dataLog.Write([]string{
			fmt.Sprintf("%d", generation),
			fmt.Sprintf("%.5f", best),
			fmt.Sprintf("%.5f", worst),
			fmt.Sprintf("%.5f", avg),
			time.Now().Format("2006-01-02 15:04:05"),
		}))
		dataLog.Flush()

		population[0].Save(fmt.Sprintf("output/gen-%010d.jpg", generation))
		// TODO: see about putting fitness on image
		population[0].Save("latest.jpg")

		fmt.Printf("Creating new population\n")
		oldPop := population
		population = Population(make([]*Individual, 0, *popSize))

		// Elitism - we keep best 10 individuals
		for i := 0; i < 10; i++ {
			population = append(population, oldPop[i])
		}

		// Now create rest of population with selection/crossover/mutation
		for len(population) < *popSize {
			// Select with tournament selection
			parent1 := Selection(oldPop)
			parent2 := Selection(oldPop)

			child1, child2 := Crossover(parent1, parent2, *crossOverRate)

			population = append(population, Mutation(child1, *mutationRate))
			population = append(population, Mutation(child2, *mutationRate))
		}
	}

	os.Exit(0)
}
