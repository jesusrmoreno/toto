package domain

// Player ..
type Player struct {
	Comm Comm
}

func (p Player) String() string {
	return p.Comm.Id()
}

// Comm provides an interface to communicate with the Player
type Comm interface {
	// Id returns the session id of socket.
	Id() string
	// Rooms returns the rooms name joined now.
	Rooms() []string
	// On registers the function f to handle an event.
	On(event string, f interface{}) error
	// Emit emits an event with given args.
	Emit(event string, args ...interface{}) error
	// Join joins the room.
	Join(room string) error
	// Leave leaves the room.
	Leave(room string) error
	// BroadcastTo broadcasts an event to the room with given args.
	BroadcastTo(room, event string, args ...interface{}) error
}

// PlayerStore ...
type PlayerStore interface {
	Store(p Player)
	Get(uuid string) (*Player, error)
}
