package main

import "math/rand"

// Crossover copies the two parents to children and performs crossover at the given rate
func Crossover(parent1 *Individual, parent2 *Individual, rate float64) (*Individual, *Individual) {
	child1 := NewIndividual(parent1.target, len(parent1.genes))
	child2 := NewIndividual(parent2.target, len(parent2.genes))

	for idx, g1 := range parent1.genes {
		g2 := parent2.genes[idx]
		if rand.Float64() <= rate {
			g1, g2 = g2, g1
		}

		child1.genes[idx] = g1.Copy()
		child2.genes[idx] = g2.Copy()
	}

	return child1, child2
}
