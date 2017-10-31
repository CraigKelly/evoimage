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

// Shuffle provides a complete shuffle of the genome (since order matters)
func Shuffle(ind *Individual) *Individual {
	clone := NewIndividual(ind.target)
	for write, read := range rand.Perm(len(clone.genes)) {
		clone.genes[write] = ind.genes[read]
	}
	return clone
}
