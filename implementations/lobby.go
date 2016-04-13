package impl

import (
	"github.com/tiltfactor/toto/domain"
)

// Lobby implements the Lobby interface in the domain
type Lobby struct {
	Queue chan domain.Player
}

// AddToQueue ...
func (l Lobby) AddToQueue(p domain.Player) {
	l.Queue <- p
}

// PopFromQueue ....
func (l Lobby) PopFromQueue() domain.Player {
	return <-l.Queue
}

// Size ...
func (l Lobby) Size() int {
	return len(l.Queue)
}
