package squid

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var random *rand.Rand // Private random

func init() {
	random = rand.New(
		&lockedRandSource{
			src: rand.NewSource(time.Now().UnixNano()),
		},
	)
}

// locked to prevent concurrent use of the underlying source
type lockedRandSource struct {
	lock sync.Mutex // protects src
	src  rand.Source
}

// to satisfy rand.Source interface
func (r *lockedRandSource) Int63() int64 {
	r.lock.Lock()
	ret := r.src.Int63()
	r.lock.Unlock()
	return ret
}

// to satisfy rand.Source interface
func (r *lockedRandSource) Seed(seed int64) {
	r.lock.Lock()
	r.src.Seed(seed)
	r.lock.Unlock()
}

func randomItem(list []string) string {
	return list[random.Intn(len(list))]
}

// GenerateID generates a new complex asana style ID
func GenerateID() string {
	count := random.Intn(10000) + 2
	firstAdjective := randomItem(Adjectives)
	noun := randomItem(SimpleSubjects)
	verb := randomItem(Verbs)
	adverb := randomItem(Adverbs)
	secondAdjective := randomItem(Adjectives)
	secondNoun := randomItem(SimpleSubjects)
	joiner := randomItem(Joiners)
	return fmt.Sprintf("%d-%s-%s-%s-%s-%s-%s-%s", count,
		firstAdjective, noun, verb, adverb, joiner, secondAdjective, secondNoun)
}

// GenerateSimpleID generates a new simple Asana style ID
func GenerateSimpleID() string {
	count := random.Intn(10000) + 2
	adj := randomItem(Adjectives)
	noun := randomItem(SimpleSubjects)
	verb := randomItem(Verbs)
	adverb := randomItem(Adverbs)
	return fmt.Sprintf("%d-%s-%s-%s-%s",
		count, adj, noun, verb, adverb)
}
