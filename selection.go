package main

import "math/rand"

// Selection assumes pop is in sorted fitness order and performs tournament selection
func Selection(pop Population) *Individual {
	winner := rand.Intn(len(pop))
	i2 := rand.Intn(len(pop))
	if i2 < winner {
		winner = i2
	}
	return pop[winner]
}
