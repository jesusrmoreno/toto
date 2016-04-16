package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/googollee/go-socket.io"

	"github.com/tiltfactor/toto/domain"
	"github.com/tiltfactor/toto/utils"

	"github.com/BurntSushi/toml"
	"github.com/jesusrmoreno/sad-squid"

	logrus "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
	log.Level = logrus.DebugLevel
}

// Version ...
var Version = "1.3.3"

// Events that are exposed to the client
const (
	connection       = "connection"
	disconnection    = "disconnection"
	playerDisconnect = "player-disconnect"
	groupAssignment  = "group-assignment"
	roomMessage      = "room-message"
	joinGame         = "join-game"
	makeMove         = "make-move"
	moveMade         = "move-made"
	inQueue          = "in-queue"

	serverError = "server-error"
	clientError = "client-error"
)

// Response is the structure of all our responses.
type Response struct {
	Timestamp int64       `json:"timeStamp"`
	Kind      string      `json:"kind"`
	Data      interface{} `json:"data"`
}

// GameJoinRequest is the request that the client should sent to get a room.
type GameJoinRequest struct {
	GameID string `json:"gameId"`
}

// Control serves to store the metadata for different games
type Control struct {
	// These must be thread safe so we use the ConcurrentMap types
	TurnMap *utils.ConcurrentStringIntMap
	// Maps the player id to the room
	RoomMap *utils.ConcurrentStringMap
}

// QueuePlayers adds players to the game's lobby to wait for a partner.
// Players are queued on a first come first serve basis.
func QueuePlayers(g domain.Game, p domain.Player) bool {
	pq := g.Lobby
	if !pq.Contains(p.Comm.Id()) {
		pq.AddToQueue(p)
		return true
	}
	return false
}

// GroupPlayers attempts to creates groups of players of the size defined in the
// game files. It also sets the player turns.
// It returns the name of the room and true if it succeeded or
// an empty string and false if it did not.
func GroupPlayers(g domain.Game, gi *Control) (string, []domain.Player) {
	log.Debug("Attempting to group players for game", g.UUID)
	pq := g.Lobby
	max := g.MaxPlayers
	min := g.MinPlayers
	if max == 0 {
		max = min
	}
	available := pq.Size()
	for needed := max; needed >= min; needed-- {
		if available >= max {
			team := []domain.Player{}
			roomName := squid.GenerateSimpleID()
			for i := 0; i < needed; i++ {
				p := pq.PopFromQueue()
				team = append(team, p)

				// Place the player in the created room.
				p.Comm.Join(roomName)

				playerID := p.Comm.Id()
				gi.RoomMap.Set(playerID, roomName)

				// We generate a turn key composed of the room name and player id to store
				// the turn. turns are assigned based off of how they are popped from the
				// queue.
				tk := TurnKey(playerID, roomName)
				gi.TurnMap.Set(tk, i)
			}
			return roomName, team
		}
	}
	return "", nil
}

// Cross origin server is used to add cross-origin request capabilities to the
// socket server. It wraps the socketio.Server
type crossOriginServer struct {
	Server *socketio.Server
}

// ServeHTTP is implemented to add the needed header for CORS in socketio.
// This must be named ServeHTTP and take the
// (http.ResponseWriter, r *http.Request) to satisfy the http.Handler interface
func (s crossOriginServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	s.Server.ServeHTTP(w, r)
}

// HandlePlayerJoin is called when a player makes a request to join a game
// it checks the validity of the passed game id and places players in the queue
// for that game if the game id is valid, it then attempts to group players.
// Because the queue and group logic is called each time a player connects
// we eliminate the need for a loop to check if there are enough players.
// Also because the Queue is protected by a mutex we don't need to worry about
// players getting assigned to multiple rooms.
func HandlePlayerJoin(so socketio.Socket, r GameJoinRequest,
	games domain.GameMap, info Control) {
	gameID := r.GameID
	if gameID == "" {
		log.Debug("No game included from", so.Id())
		so.Emit(clientError, ErrorResponse(clientError, "Must include GameID"))
	}
	log.Debug(so.Id(), "attempting to join game", gameID)
	// If the player attempts to connect to a game we first have to make
	// sure that they are joining a game that is registered with our server.
	if g, exists := games[gameID]; exists {
		// First queue the player
		newPlayer := domain.Player{
			Comm: so,
		}
		if didQueue := QueuePlayers(g, newPlayer); didQueue {
			// Create the response we're going to send
			r := WrapResponse(inQueue, struct {
				Msg            string `json:"message"`
				PlayersInQueue int    `json:"playersInQueue"`
			}{
				Msg:            "You are in the queue for game: " + g.Title,
				PlayersInQueue: g.Lobby.Size(),
			})
			so.Emit(inQueue, r)
			if rn, group := GroupPlayers(g, &info); group != nil && rn != "" {
				// Tell each member what their room name is as well as their turn
				for i, p := range group {
					data := map[string]interface{}{}
					data["roomName"] = rn
					data["turnNumber"] = i
					r := WrapResponse(groupAssignment, data)
					p.Comm.Emit(groupAssignment, r)
				}
			}
		} else {
			// Create the response we're going to send
			data := map[string]interface{}{}
			data["message"] = "Already in queue"
			r := WrapResponse(clientError, data)
			so.Emit(clientError, r)
		}
	} else {
		log.Debug("Invalid GameId from", so.Id())
		so.Emit(clientError, ErrorResponse(clientError, "Invalid GameID"))
	}
}

// StartServer loads the games from the games directory (exits on error)
// Creates the socket io server and wraps it to accept all origins
// Initializes our Control structure to store metadata
// and finally starts up the socket io server.
func StartServer(c *cli.Context) {
	games, err := ReadGameFiles("./games")
	if err != nil {
		log.Fatal(err)
	}
	for key, game := range games {
		log.Println("Loaded:", key, "from", game.FileName)
	}
	server, err := socketio.NewServer(nil)
	s := crossOriginServer{
		Server: server,
	}
	if err != nil {
		log.Fatal(err)
	}
	info := Control{
		RoomMap: utils.NewConcurrentStringMap(),
		TurnMap: utils.NewConcurrentStringIntMap(),
	}

	server.On(connection, func(so socketio.Socket) {
		log.Debug("Connection from", so.Id())

		// Makes it so that the player joins a room with his/her unique id.
		so.Join(so.Id())
		so.On(joinGame, func(r GameJoinRequest) {
			HandlePlayerJoin(so, r, games, info)
		})

		so.On(disconnection, func() {
			// This is really really bad unfuture proof, slow code.
			// Please Refactor me
			for key := range games {
				g := games[key]
				g.Lobby.Remove(so.Id())
			}
			r, foundRoom := info.RoomMap.Get(so.Id())
			tk := TurnKey(so.Id(), r)
			t, foundTurn := info.TurnMap.Get(tk)
			// Broadcast to the room that the player disconnected.
			if foundRoom && foundTurn {
				m := map[string]interface{}{}
				m["player"] = t
				server.BroadcastTo(r, playerDisconnect, WrapResponse(playerDisconnect, m))
			}
			// Remove the player from the room and turn maps when they disconnect.
			info.RoomMap.Del(so.Id())
			info.TurnMap.Del(tk)
		})

		so.On(makeMove, func(move json.RawMessage) {
			room, exists := info.RoomMap.Get(so.Id())
			log.Println(string(move))
			if exists {
				m := map[string]interface{}{}
				if err := json.Unmarshal(move, &m); err != nil {
					log.Debug("Invalid JSON from", so.Id(), string(move))
					so.Emit(clientError, ErrorResponse(clientError, "Invalid JSON"))
				}
				turn, exists := info.TurnMap.Get(room + ":" + so.Id())
				if !exists {
					log.Debug("No turn assigned", so.Id())
					so.Emit(serverError, ErrorResponse(serverError, "No turn assigned"))
				}
				// Overwrites who's turn it is using the turn map assigned at join.
				m["madeBy"] = turn
				m["madeById"] = so.Id()
				r := WrapResponse(moveMade, m)
				log.Println(r)
				server.BroadcastTo(room, moveMade, r)
			} else {
				log.Debug("No room assigned for", so.Id())
				so.Emit(serverError, ErrorResponse(serverError, "Not in any Room"))
			}
		})
	})

	port := c.String("port")

	http.Handle("/socket.io/", s)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// WrapResponse wraps the data we want to send in our response struct and adds
// the Timestamp and kind of response.
func WrapResponse(kind string, data interface{}) Response {
	return Response{
		Timestamp: time.Now().UnixNano(),
		Kind:      kind,
		Data:      data,
	}
}

// ErrorResponse is a method for creating errors more quickly.
// It takes the error string and then calls WrapResponse internally to wrap the
// data
func ErrorResponse(kind, err string) Response {
	d := map[string]interface{}{}
	d["error"] = err
	return WrapResponse(kind, d)
}

// ReadGameFiles reads the provided directory for files that conform to the
// game struct definition, these must be json files, and loads them into our
// game map.
func ReadGameFiles(gameDir string) (domain.GameMap, error) {
	files, err := filepath.Glob(gameDir + "/*.toml")
	if len(files) == 0 {
		return nil, errors.New("Unable to find games. Does games directory exist?")
	}

	gm := domain.GameMap{}
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		raw, err := os.Open(f)
		defer raw.Close()
		if err != nil {
			return nil, err
		}
		r := io.Reader(raw)
		dummy := domain.Game{}
		if meta, err := toml.DecodeReader(r, &dummy); err != nil {
			log.Error(meta)
			return nil, errors.New("Invalid configuration in file: " + f)
		}
		if dummy.MinPlayers == 0 {
			return nil, errors.New("Invalid configuration in file: must provide minPlayers" + f)
		}
		g := domain.Game{
			MinPlayers: dummy.MinPlayers,
			MaxPlayers: dummy.MaxPlayers,
			Title:      dummy.Title,
			UUID:       dummy.UUID,
			Lobby:      domain.NewLobby(),
		}
		g.FileName = f
		if _, exists := gm[g.UUID]; exists {
			return nil, errors.New("uniqueKey conflict between: " + f + " and " +
				gm[g.UUID].FileName)
		}
		gm[g.UUID] = g
	}
	return gm, nil
}

// TurnKey returns the generated key for storing turns
func TurnKey(playerID, roomName string) string {
	return roomName + ":" + playerID
}

func main() {
	app := cli.NewApp()
	app.Name = "Toto"
	app.Usage = "a server for creating quick prototype websocket based games."
	app.Action = StartServer
	app.Version = Version
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "port, p",
			Value: "3000",
			Usage: "The port to run the server on",
		},
	}
	app.Run(os.Args)
}
