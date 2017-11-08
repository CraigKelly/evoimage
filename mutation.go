package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"
)

// Given a source value and a stddev, return a mutated number
// - insure that the abs val of the delta is at least 1
// - insure that the returned value is clamped to [mn,mx]
func mutateNorm(src float64, sd float64, mn float64, mx float64) float64 {
	d := rand.NormFloat64() * sd
	if math.Abs(d) < 1.0 {
		if d < 0.0 {
			d = -1.0
		} else {
			d = 1.0
		}
	}

	newval := src + d
	if newval < mn {
		newval = mn
	} else if newval > mx {
		newval = mx
	}

	return newval
}

// Given a color coord (RGB), return a mutated coord
func mutateColorCoord(c uint8) uint8 {
	return uint8(mutateNorm(float64(c), 5.0, 0.0, 255.0))
}

// Mutation returns a mutated individual: WHICH IS CURRENTLY INPLACE
func Mutation(ind *Individual, rate float64) *Individual {
	var clr *color.NRGBA

	// We can precompute these
	lim := ind.target.imageData.Bounds()
	mnx, mny := float64(lim.Min.X), float64(lim.Min.Y)
	mxx, mxy := float64(lim.Max.X), float64(lim.Max.Y)

	mutatePoint := func(p image.Point) image.Point {
		p.X = int(mutateNorm(float64(p.X), 5.0, mnx, mxx))
		p.Y = int(mutateNorm(float64(p.Y), 5.0, mny, mxy))
		return p
	}

	for _, curr := range ind.genes {
		// colors
		clr = curr.destColor
		if rand.Float64() <= rate {
			clr.R = mutateColorCoord(clr.R)
		}
		if rand.Float64() <= rate {
			clr.G = mutateColorCoord(clr.G)
		}
		if rand.Float64() <= rate {
			clr.B = mutateColorCoord(clr.B)
		}
		if rand.Float64() <= rate {
			clr.A = mutateColorCoord(clr.A)
		}

		// vertices
		for idx := range curr.destVertices {
			if rand.Float64() <= rate {
				curr.destVertices[idx] = mutatePoint(curr.destVertices[idx])
			}
		}
	}

	return ind
}

// Shuffle provides a complete shuffle of the genome (since order matters)
func Shuffle(ind *Individual) *Individual {
	clone := NewIndividual(ind.target)
	for write, read := range rand.Perm(len(clone.genes)) {
		clone.genes[write] = ind.genes[read].Copy()
	}
	return clone
}
