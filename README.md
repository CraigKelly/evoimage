# evoimage

Simple attempt at evolving image reproductions

## Overview

This project is experimental and for fun. It doesn't really accomplish anything
that is publication worthy, but you might enjoy playing with it.  See `main.go`
for command line options and the main loop. See `representation.go` for the
main representation and encoding.

## Implementation

The main loop is a standard generational GA, but includes some adaptations
during the run. (Note that we equate "making progress" with an increase in the
best individual - which is a decrease is our minimized fitness score).

* We use elitism (see Elitism below)
* Mutation rate is increased when progress has stalled (see Mutation below)
* The population size increases every generation progress is not made, but returns
  to the default level when progress resumes
* Tournament size is rotated (see Selection below)

## Fitness Function

The fitness function is the sum of the Euclidean distance in RGB space for all
pixels. Note that this means we are attempting to *minimize* our fitness
function.

See `representation.go`.

## Representation

Each individual is an ordered list of genes where each gene is a triangle and
a color in RGBA space (note our use of transparency for representation and
drawing, but not in the final fitness function).

Color and spatial coordinates are sampled uniformly at random when creating a
random instance.

See `representation.go`.

## Selection

Selection is currently tournament selection. The main loop uses a rotating
tournament size (2-5 inclusive).

See `selection.go` and `main.go`.

## Mutation

We use Gaussian mutation. Currently the standard deviation is different for
spatial and color coordinates and is hand-coded.

The mutation rate currently receives a (capped) increase for every generation
since we have failed to get an increase in the best fitness score.

There is also a gene shuffle operator used as part of our elitism strategy (see
below).

See `mutation.go`.

## Crossover

Crossover is uniform crossover.

See `crossover.go`.

## Elitism

We copy the five best individuals to the next generation.

We also add a *second* copy of the best five individuals, but with their genes
shuffled and mutated.

## Installing

You can just use the executable in the folder matching your operating system
and architecture under dist.

## Building and running

To build this project, you need Go 1.9 and recent version of GNU Make running
in a bash shell (modern bash is preferred). All builds and runs have been
performed on Ubuntu 16.04 and 17.04, so everything you need can be installed
via apt. However, please note that you will need to install Go 1.9 (or later)
manually. You can use ubuntu-make for this - see
https://github.com/ubuntu/ubuntu-make

Go is cross-platform, so you should be able to build on Windows or Mac OSX, bu
that has not been tested.

Go dependencies are managed with `dep` (see https://github.com/golang/dep for
details), but you shouldn't need to worry about that unless you are adding or
upgrading dependencies.

See `Makefile`. The short version is that static source code analysis and
formatting can be checked with `make lint`. Build with `make`.

After building, run `./evoimage -h` to see all parameter options.

To run on an image with all default parameters, you only need to supply the
`-image` parameter.  For example, `./evoimage -image imgs/target-mondrian.jpg`. 

`evoimage` appends to log files (named for the target image) in the `log`
directory during the run.  It also writes the best image for each generation to
the `output` directory. Image name formats use the generation number but not
the file name, so you should save or clear the output directory before starting
a new run. We also write the best image as `./latest.jpg`.

Running `./script/output_ani` will take all current images in the output
directory and create an mp4 video showing progress. Note that `ffmpeg` must be
installed.

## Images

* "Portrait of Isabel Parreño y Arce, Marquesa de Llano, Anton Raphael Mengs, 1771 - 1772"
  by Anton Raphael Mengs is licensed under CC0 1.0
  https://ccsearch.creativecommons.org/image/detail/TJmttlM53HMeELT3IduwGw==
* "Landscape" by Carle (Antoine Charles Horace) Vernet (French, Bordeaux 1758–1836 Paris)
  via The Metropolitan Museum of Art is licensed under CC0 1.0 
  https://ccsearch.creativecommons.org/image/detail/7J6dR75Zr50Kwq7_Ax0ehw==
* "Composition" by Piet Mondrian (Dutch, Amersfoort 1872–1944 New York) via
  The Metropolitan Museum of Art is licensed under CC0 1.0 
  https://ccsearch.creativecommons.org/image/detail/wGM3NBEVzE0rIfI48n1NIQ==
* "mondrian" by apenny is licensed under CC BY 2.0 
  https://ccsearch.creativecommons.org/image/detail/dci6KvaAj9FaCqASttXsEQ==
* "Versailles" by Auguste Renoir (French, Limoges 1841–1919 Cagnes-sur-Mer) via
  The Metropolitan Museum of Art is licensed under CC0 1.0
  https://ccsearch.creativecommons.org/image/detail/LZKw8dOJS85mFqXEMeWrwA==
* "The Milliner" by Auguste Renoir (French, Limoges 1841–1919 Cagnes-sur-Mer) via
  The Metropolitan Museum of Art is licensed under CC0 1.0 
  https://ccsearch.creativecommons.org/image/detail/6DhumcJp7g8tXzWCgIfjog==

The original full size CC JPEG's from above are in the `imgs` directory with
the prefix `orig-`.  They have been scaled down so the largest dimension is 256
pixels while preserving the original aspect ratio.  The scaled images used as
target images are also in the `imgs` directory with the prefix `target`.

