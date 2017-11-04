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
	"path/filepath"
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

	work := make(chan int, 256)

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
	mutationRate := flags.Float64("mutationRate", 0.08, "Mutation rate to use")
	crossOverRate := flags.Float64("crossoverRate", 0.60, "Crossover rate to use")
	popSize := flags.Int("popSize", 200, "Population size in a generation")
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

	_, imageBase := filepath.Split(*image)
	logFileName := fmt.Sprintf("logs/%s-log.csv", imageBase)
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

	tournSize := 1 // special: first tournament size will be two
	lastBest := float64(0.0)
	stallCount := 0
	adaptMutRate := *mutationRate
	adaptPopSize := *popSize

	for generation := 0; generation < 100000; generation++ {
		// Image creation and evaluation across all cores
		evalPop(population, cores)

		// Now we can sort and find best/worst
		sort.Sort(population)
		best := population[0].Fitness()
		worst := population[len(population)-1].Fitness()
		avg := population.MeanFitness()

		if math.Abs(best-lastBest) < 0.0000001 {
			stallCount++
		} else {
			stallCount = 0
		}
		lastBest = best

		tournSize++
		if tournSize > 5 {
			tournSize = 2
		}

		adaptMutRate = *mutationRate + (0.015 * float64(stallCount))
		if adaptMutRate > 0.25 {
			adaptMutRate = 0.25
		}

		adaptPopSize = *popSize + (stallCount * 2)

		pcheck(dataLog.Write([]string{
			fmt.Sprintf("%d", generation),
			fmt.Sprintf("%.5f", best),
			fmt.Sprintf("%.5f", worst),
			fmt.Sprintf("%.5f", avg),
			time.Now().Format("2006-01-02 15:04:05"),
		}))
		dataLog.Flush()

		log.Printf(
			"Gen:%5d PS:%5d SC:%d,TS:%d,MR:%.5f best %.2f <=> avg %.2f <=> worst %.2f\n",
			generation, len(population),
			stallCount, tournSize, adaptMutRate,
			best, avg, worst,
		)

		population[0].Save(fmt.Sprintf("output/gen-%010d.jpg", generation))
		population[0].Save("latest.jpg")

		oldPop := population
		population = Population(make([]*Individual, 0, adaptPopSize+5+(stallCount/2)))

		// Elitism - we keep best 5 individuals AND a shuffled copy of the best 5
		for i := 0; i < 5; i++ {
			population = append(population, oldPop[i])
			population = append(population, Shuffle(oldPop[i]))
		}

		// Now create rest of population with selection/crossover/mutation
		for len(population) < adaptPopSize {
			// Select with tournament selection
			parent1 := Selection(oldPop, tournSize)
			parent2 := Selection(oldPop, tournSize)

			child1, child2 := Crossover(parent1, parent2, *crossOverRate)

			population = append(population, Mutation(child1, adaptMutRate))
			population = append(population, Mutation(child2, adaptMutRate))
		}

		// Inject randomness if we are stalled
		for i := 0; i < (stallCount / 2); i++ {
			ind := NewIndividual(target)
			ind.RandInit()
			population = append(population, ind)
		}
	}

	os.Exit(0)
}
