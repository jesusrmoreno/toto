package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/googollee/go-socket.io"

	"github.com/jesusrmoreno/toto/domain"
	"github.com/jesusrmoreno/toto/implementations"
	"github.com/jesusrmoreno/toto/utils"

	"github.com/BurntSushi/toml"
	"github.com/jesusrmoreno/sad-squid"

	"github.com/codegangsta/cli"
)

// Version ...
var Version = "No Version Specified"

// Events that are exposed to the client
const (
	groupAssignment = "group-assignment"
	roomMessage     = "room-message"
	leaveGroup      = "leave-group"
	joinGroup       = "join-group"
	joinGame        = "join-game"
	getPeers        = "get-peers"
	makeMove        = "make-move"
	moveMade        = "move-made"
	inQueue         = "in-queue"

	serverError = "server-error"
	clientError = "client-error"
)

// Response ...
type Response struct {
	Kind string                 `json:"kind"`
	Data map[string]interface{} `json:"data"`
}

// Method for creating errors more quickly.
func errorResponse(kind, err string) Response {
	d := map[string]interface{}{}
	d["error"] = err
	r := Response{
		Kind: kind,
		Data: d,
	}
	return r
}

// GamesInfo serves to store the metadata for different games
type GamesInfo struct {
	// These must be thread safe so we use the ConcurrentMap types
	TurnMap *utils.ConcurrentStringIntMap
	// Maps the player id to the room
	RoomMap *utils.ConcurrentStringMap
}

// ReadGameFiles reads the provided directory for files that conform to the
// game struct definition, these must be json files, and loads them into our
// game map.
func ReadGameFiles(gameDir string) (domain.GameMap, error) {
	files, err := filepath.Glob(gameDir + "/*.toml")
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
			log.Println(meta)
			return nil, errors.New("Invalid configuration in file: " + f)
		}
		g := domain.Game{
			Players: dummy.Players,
			Title:   dummy.Title,
			UUID:    dummy.UUID,
			Lobby: impl.Lobby{
				// We implement our lobby as a buffered channel that takes the number of
				// players specified in the config file for the game and makes that the
				// cap.
				Queue: make(chan domain.Player, dummy.Players),
			},
		}
		g.FileName = f
		if _, exists := gm[g.UUID]; exists {
			return nil, errors.New("uniqueKey conflict between: " + f + " and " + gm[g.UUID].FileName)
		}
		gm[g.UUID] = g
	}
	return gm, nil
}

// QueuePlayers adds players to the game's lobby to wait for a partner.
// Players are queued on a first come first serve basis.
func QueuePlayers(g domain.Game, p domain.Player) {
	data := map[string]interface{}{}
	data["message"] = "You are in the queue for game: " + g.Title
	r := Response{
		Kind: inQueue,
		Data: data,
	}
	p.Comm.Emit(inQueue, r)
	pq := g.Lobby
	pq.AddToQueue(p)
}

// GroupPlayers creates groups of players of size PlayerSize as defined in the
// game files.
func GroupPlayers(g domain.Game, server *socketio.Server, gi *GamesInfo) {
	pq := g.Lobby
	n := g.Players
	for {
		select {
		default:
			if g.Lobby.Size() >= n {
				team := []domain.Player{}
				// Generate a room id
				roomName := squid.GenerateSimpleID()
				for i := 0; i < n; i++ {
					p := pq.PopFromQueue()
					pID := p.Comm.Id()
					team = append(team, p)

					// Set the metadata
					gi.RoomMap.Set(pID, roomName)
					gi.TurnMap.Set(roomName+":"+pID, i)

					// Place the player in the created room.
					p.Comm.Join(roomName)

					// Create the response
					data := map[string]interface{}{}
					data["roomName"] = roomName
					data["turnNumber"] = i
					r := Response{
						Kind: inQueue,
						Data: data,
					}
					p.Comm.Emit(groupAssignment, r)
				}

				// Create the response
				data := map[string]interface{}{}
				data["message"] = fmt.Sprintf("Welcome to %s", roomName)
				r := Response{
					Kind: roomMessage,
					Data: data,
				}
				server.BroadcastTo(roomName, roomMessage, r)
			}
		}
	}
}

func (s customServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	origin := r.Header.Get("Origin")
	w.Header().Set("Access-Control-Allow-Origin", origin)
	s.Server.ServeHTTP(w, r)
}

type customServer struct {
	Server *socketio.Server
}

// StartServer ...
func StartServer(c *cli.Context) {
	games, err := ReadGameFiles("./games")
	for key, game := range games {
		log.Println("Loaded:", key, "from", game.FileName)
	}
	if err != nil {
		log.Fatal(err)
	}

	server, err := socketio.NewServer(nil)
	s := customServer{
		Server: server,
	}

	if err != nil {
		log.Fatal(err)
	}

	info := GamesInfo{
		RoomMap: utils.NewConcurrentStringMap(),
		TurnMap: utils.NewConcurrentStringIntMap(),
	}
	for key := range games {
		go GroupPlayers(games[key], server, &info)
	}
	server.On("connection", func(so socketio.Socket) {
		// Makes it so that the player joins a room with his/her unique id.
		so.Join(so.Id())
		log.Println("Connection from", so.Id())
		so.On(joinGame, func(req json.RawMessage) {
			m := map[string]string{}
			if err := json.Unmarshal(req, &m); err != nil {
				log.Println("Invalid JSON from", so.Id())
				so.Emit(clientError, errorResponse(clientError, "Invalid JSON"))
			}
			gameID, exists := m["gameId"]
			if !exists {
				log.Println("No game id from", so.Id())
				so.Emit(clientError, errorResponse(clientError, "Must include GameID"))
			}
			log.Println(so.Id(), "attempted to join game", gameID)
			// If the player attempts to connect to a game we first have to make
			// sure that they are joining a game that is registered with our server.
			if g, exists := games[gameID]; exists {
				QueuePlayers(g, domain.Player{
					Comm: so,
				})
			} else {
				log.Println("Invalid GameId from", so.Id())
				so.Emit(clientError, errorResponse(clientError, "Invalid GameID"))
			}
		})
		so.On(makeMove, func(move json.RawMessage) {
			room, exists := info.RoomMap.Get(so.Id())
			if exists {
				m := map[string]interface{}{}
				if err := json.Unmarshal(move, &m); err != nil {
					log.Println("Invalid JSON from", so.Id())
					so.Emit(clientError, errorResponse(clientError, "Invalid JSON"))
				}
				turn, exists := info.TurnMap.Get(room + ":" + so.Id())
				if !exists {
					log.Println("No turn assigned", so.Id())
					so.Emit(serverError, errorResponse(serverError, "No turn assigned"))
				}
				// Overwrites who's turn it is using the turn map assigned at join.
				m["madeBy"] = turn
				m["madeById"] = so.Id()
				r := Response{
					Kind: moveMade,
					Data: m,
				}
				so.BroadcastTo(room, moveMade, r)
			} else {
				log.Println("No room assigned for", so.Id())
				so.Emit(serverError, errorResponse(serverError, "Not in any Room"))
			}
		})
	})

	port := c.String("port")

	http.Handle("/socket.io/", s)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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
