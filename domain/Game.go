package domain

import "sync"

// GameMap serves as an in memory store of the different registered games
type GameMap map[string]Game

// Lobby is basically a FIFO queue that is threadsafe through the usage of mutex
// it uses a slice as the data store and a string to bool map to keep of track
// of items in the queue
type Lobby struct {
	Protect  *sync.RWMutex
	data     []Player
	contains map[string]bool
}

// NewLobby instantiates a new lobby (queue)
func NewLobby() *Lobby {
	return &Lobby{
		Protect:  &sync.RWMutex{},
		data:     []Player{},
		contains: make(map[string]bool),
	}
}

// AddToQueue adds a player to the queue
func (l *Lobby) AddToQueue(p Player) {
	l.Protect.Lock()
	defer l.Protect.Unlock()
	l.data = append(l.data, p)
	l.contains[p.Comm.Id()] = true
}

// PopFromQueue returns a player from the queue: FIFO
func (l *Lobby) PopFromQueue() Player {
	l.Protect.Lock()
	defer l.Protect.Unlock()
	item, a := l.data[0], l.data[1:]
	l.data = a
	delete(l.contains, item.Comm.Id())
	return item
}

// Size returns the number of items in the q
func (l *Lobby) Size() int {
	l.Protect.Lock()
	defer l.Protect.Unlock()
	return len(l.data)
}

// Contains returns true if the q contains the player with the given id
func (l *Lobby) Contains(id string) bool {
	l.Protect.Lock()
	defer l.Protect.Unlock()
	_, exists := l.contains[id]
	return exists
}

// Remove removes the player with the specified id
func (l *Lobby) Remove(id string) {
	l.Protect.Lock()
	defer l.Protect.Unlock()

	b := l.data[:0]
	for _, x := range l.data {
		if x.Comm.Id() != id {
			b = append(b, x)
		}
	}
	l.data = b
}

// Game contains all of our registered game information
type Game struct {
	FileName string `toml:"-"`
	Lobby    *Lobby
	Players  int    `toml:"playersInGroup"`
	Title    string `toml:"displayTitle"`
	UUID     string `toml:"uniqueKey"`
}
