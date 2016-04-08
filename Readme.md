# Exposed Events Draft
### join-game
join-game should be called with a uniqueKey that corresponds to the game the
player wishes to join. This will place the player on a queue until the amount of
players defined in minPlayersInGroup are on the queue. When that happens a unique
room id will be generated and the assigned-room is emitted to the player with
the room name.

### join-group
Should be used by the player to join an existing group. A player can only join
an existing group if doing so will not cause the maxPlayersInGroup limit to be
exceeded.

### get-peers
This returns a list of the members in the current room. This should be used for
figuring out things like turns etc..

### leave-group
Removes the player from the group, if the players in the room falls below
minPlayersIngGroup then the room will close.

### player-make-move
This is sent from the client to the server when the player wants to make a move,
the "move" is to be sent as a string value which should then be parsed as json
by the other clients. The server currently makes no attempt to validate the json
string but future iteration may include such a feature. The json string is
therefore sent to all players in the group and it is up to the client to figure
out who is to do what.

### player-move
this is the event that clients should listen to in order to receive moves that
other players have made. the moves should be a string version of a json object
that can be parsed by the client but the server makes no promises that such a
string is passed. As such the client should never execute anything that comes
from this event

### client-error
This should be sent by clients to the server in the case of a client error,
this can be anything from malformed json to a bad move attempt on behalf of a
different client. This is broadcast to all members of the group.

### server-error
This is sent by the server when an error occurs that generates on the server.
An example is attempting to join a gameID for which there is no json file
defined.
