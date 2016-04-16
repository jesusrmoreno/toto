# Toto :video_game:
__Toto__ is a game server for creating quick prototypes of online multiplayer
games. By leveraging socket.io and exposing a few key events to act essentially
as a message relay. Toto makes it easy to start prototyping multiplayer games
without worrying about writing a custom server for each prototype.

__Toto__ __*is not*__ meant to be a finalized game server.
You'll notice that __Toto__ offers nothing in the way of state management or
validation this is by design because __Toto__ is meant to be a prototyping
server and nothing more. As a result it assumes perfect and well behaved
clients.

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

# Events and JSON structure.
```javascript
// To join the game clickRace we set the gameId to the uniqueKey defined in
// the game file
{
  "gameId": "clickRace"
}

// If the gameId does not exist then you will receive a client error
// client-error and server-error will always include the error field
{
  "timeStamp": 1460792552456716300,
  "kind": "client-error",
  "data": {
    "error": "Invalid GameID"
  }
}



// You will then receive a in-queue message
// in-queue will include the message and the number of players currently in the
// queue.
{
  "timeStamp": 1460792552507366000,
  "kind": "in-queue",
  "data": {
    "message": "You are in the queue for game: Click Race!",
    "playersInQueue": 1
  }
}

// After a while you will receive the group-assignment message
// group-assignment will always include the room name and the turn number
// assigned to the client
{
  "timeStamp": 1460792555406774000,
  "kind": "group-assignment",
  "data": {
    "roomName": "2068-upset-pigs-swam-reproachfully",
    "turnNumber": 0
  }
}

// At this point the player has been added to the group room by the server so
// we can make a move by emitting the event "make-move". This JSON could contain anything, 
// with two exceptions. 
// The fields madeBy and madeById will be overwritten by the server with the associated 
// socketid and turn number.
{
    "clicks": 1
}

// All clients will receive the move 
// And it is up to the client to only apply the other players moves
// In this way you can also confirm that your move was indeed sent.
{
  "timeStamp": 1460792555410103300,
  "kind": "move-made",
  "data": {
    "clicks": 1,
    "madeBy": 0,
    "madeById": "RazcS5nrgT-2G7kX4HPP"
  }
}

// If a player disconnects then the player-disconnect event will be emitted.
// player-disconnect will have the turn number of the player who disconnected
{
  "timeStamp": 1460792704214456000,
  "kind": "player-disconnect",
  "data": {
    "player": 1
  }
}
```
