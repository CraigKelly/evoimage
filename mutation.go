package main

import "math/rand"

// Mutation returns a mutated individual: WHICH IS CURRENTLY INPLACE
func Mutation(ind *Individual, rate float64) *Individual {
	for i := range ind.genes {
		if rand.Float64() <= rate {
			ind.genes[i] = NewGene(ind.target)
		}
	}
	return ind
}
