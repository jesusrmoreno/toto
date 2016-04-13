package domain

// GameMap serves as an in memory store of the different registered games
type GameMap map[string]Game

// Game contains all of our registered game information
type Game struct {
	FileName string `toml:"-"`
	Lobby    *Lobby
	Players  int    `toml:"playersInGroup"`
	Title    string `toml:"displayTitle"`
	UUID     string `toml:"uniqueKey"`
}
