package domain

import "encoding/json"

// GameMap serves as an in memory store of the different registered games
type GameMap map[string]Game

// Lobby ...
type Lobby interface {
	AddToQueue(p Player)
	PopFromQueue() (p Player)
	Size() int
}

// Game contains all of our registered game information
type Game struct {
	FileName       string `json:"-"`
	Lobby          Lobby
	Players        int             `json:"playersInGroup"`
	Title          string          `json:"displayTitle"`
	UUID           string          `json:"uniqueKey"`
	ExamplePayload json.RawMessage `json:"examplePayload"`
}

// ValidateMessage will eventually validate the messages passed between the
// clients but it is not implemented right now.
func (g Game) ValidateMessage(pl string) (bool, error) {
	return true, nil
}
