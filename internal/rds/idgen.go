// Copyright (C) 2026 Alexey Zapparov
// SPDX-License-Identifier: AGPL-3.0-or-later

package rds

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// Incomplete collection of ska legends across the waves.
//
// Some bands got cut not from lack of love, but because names were too spicy
// for demoing in front of suits, too long for a proper sleeve, or ended with
// numbers. Got a favourite? Open a PR!
var idGenPrefixes = []string{
	// 1st wave
	"derrick-morgan",
	"desmond-dekker",
	"don-drummond",
	"ethiopians",
	"heptones",
	"jackie-mittoo",
	"laurel-aitken",
	"paragons",
	"prince-buster",
	"roland-alphonso",
	"roy-ellis",
	"skatalites",
	"the-gaylads",
	"the-maytals",
	"the-pioneers",
	"the-upsetters",
	"tommy-mccook",
	// 2nd wave
	"bad-manners",
	"judge-dread",
	"madness",
	"the-beat",
	"the-bodysnatchers",
	"the-selecter",
	"the-specials",
	// 3rd wave
	"buck-o-nine",
	"fishbone",
	"goldfinger",
	"hepcat",
	"less-than-jake",
	"mustard-plug",
	"no-doubt",
	"operation-ivy",
	"pietasters",
	"rancid",
	"reel-big-fish",
	"rx-bandits",
	"save-ferris",
	"skankin-pickle",
	"slackers",
	"sublime",
	"the-aquabats",
	"the-toasters",
	// Other beautiful noise
	"bestas",
	"kortatu",
	"skontra",
	"trapallada",
}

// idGen is like legendary Jordi Panico providing a perfectly random tune
// and keeping count of beers served. Thread-safe, so no matter how many
// rudeboys trample each other rushing the counter, each gets their own brew.
type idGen struct {
	mu      sync.Mutex
	counter int
	rng     *rand.Rand
}

// Generate pulls the next record from the crate -- random band, fresh number.
// Stomp your [Query.Label] with it and stop your messing around, Rudy!
func (g *idGen) Generate() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.counter++
	prefix := idGenPrefixes[g.rng.IntN(len(idGenPrefixes))]

	return fmt.Sprintf("%s-%04d", prefix, g.counter)
}

// NewIDGeneratorWithSeed is the soundcheck version -- lets you tune
// the RNG to a specific frequency, when you need your setlist to be
// reproducible, not random like a real punk show.
func NewIDGeneratorWithSeed(seed uint64) *idGen {
	return &idGen{
		rng: rand.New(rand.NewPCG(seed, seed)),
	}
}

// NewIDGenerator sets up a fresh generator seeded with the current time.
// Ready to rock from the moment you plug it in.
func NewIDGenerator() *idGen {
	return NewIDGeneratorWithSeed(uint64(time.Now().UnixNano()))
}
