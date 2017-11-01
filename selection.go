package main

import "math/rand"

// Selection assumes pop is in sorted fitness order and performs tournament selection
func Selection(pop Population, tournSize int) *Individual {
	winner := rand.Intn(len(pop))

	for i := 1; i < tournSize; i++ {
		contender := rand.Intn(len(pop))
		if contender < winner {
			winner = contender
		}
	}

	return pop[winner]
}
