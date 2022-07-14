package datagen

import (
	"math/rand"
	"testing"
	"time"
)

func AddRandomSeedsToFuzzer(f *testing.F, num uint) {
	// Seed based on the current time
	rand.Seed(time.Now().Unix())
	var idx uint
	for idx = 0; idx < num; idx++ {
		f.Add(rand.Int63())
	}
}
