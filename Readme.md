# Toto :video_game:
__Toto__ is a game server for creating quick prototypes of websocket based
games. By exposing a few key events Toto makes it easy to start making
multiplayer games. Toto serves as a relay point for communication between
clients. It does not handle any of the game logic and expects that the clients
will check all relayed messages.

# Installing
```bash
go get github.com/tiltfactor/toto
```

# Running
```bash
# To run on port 8080
toto --port 8080

# To run on default port (3000)
toto
```

# Upgrading
```bash
go get -u github.com/tiltfactor/toto
```


# How
First a game definition must be created and placed in a folder called _games_ in
the directory where Toto is run. A game definition is written in TOML and
looks like this:
```toml
# The number of players required to create a group
minPlayers = 2

# Optional maxPlayer
maxPlayers = 2

# The title that will be displayed should be displayed to the user
displayTitle = "This is an example game!"

# A unique key that the client will use to request to join a game
uniqueKey = "example-game"
```

Here players in group states the number of players that must be placed in a
group.
Display title is what players will be shown, and uniqueKey is the key that
__Toto__ expects the client to send when requesting to join a game.
These are all required fields and __Toto__ will throw an error if there are
fields missing or if the uniqueKey conflicts with another declared game.

# Events
### join-game
join-game should be called with a uniqueKey that corresponds to the game the
player wishes to join. This will place the player on a queue until the amount of
players defined in playersInGroup are on the queue. When that happens a unique
room id will be generated and the assigned-room is emitted to the player with
the room name.

### player-disconnect
Will be sent in the event of a player disconnecting from the group. The player's
turn will be included.

### in-queue
Sent when the player first gets added to the game queue.

### group-assignment
Sends the player the room name and turn assigned. The client should use this to
initialize the client turn and room name.

### room-message
Is sent to everyone in the room. The client should subscribe to this to get info
about room events.

### make-move
This is sent from the client to the server when the player wants to make a move,
the "move" is to be sent as a json object that is then sent to be used
by the other clients. The server currently makes no attempt to validate the json but future iteration 
may include such a feature. The json is therefore sent to all players in the group and it is up to the client to figure
out who is to do what.

### move-made
this is the event that clients should listen to in order to receive moves that
other players have made. the moves should be a json object
that can be parsed by the client but the server makes no promises that such a
string is parsed or safe. As such the client should never execute (eval) anything that comes
from this event

### client-error
This should be sent by clients to the server in the case of a client error,
this can be anything from malformed json to a bad move attempt on behalf of a
different client. This is broadcast to all members of the group.

### server-error
This is sent by the server when an error occurs that generates on the server.
An example is attempting to join a gameID for which there is no json file
defined.
