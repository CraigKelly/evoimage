package main

import "math/rand"

// Crossover copies the two parents to children and performs crossover at the given rate
func Crossover(parent1 *Individual, parent2 *Individual, rate float64) (*Individual, *Individual) {
	child1 := NewIndividual(parent1.target)
	child2 := NewIndividual(parent2.target)

	for idx, g1 := range parent1.genes {
		g2 := parent2.genes[idx]
		if rand.Float64() <= rate {
			g1, g2 = g2, g1
		}

		var copy1, copy2 Gene
		copy1 = *g1
		copy2 = *g2

		child1.genes[idx] = &copy1
		child2.genes[idx] = &copy2
	}

	return child1, child2
}
