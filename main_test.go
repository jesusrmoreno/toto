package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/tiltfactor/toto/domain"
)

type testComm struct {
	ID string
}

func (t testComm) Id() string {
	return t.ID
}

func (t testComm) Rooms() []string {
	return []string{}
}

func (t testComm) On(event string, f interface{}) error {
	return nil
}

func (t testComm) Emit(event string, args ...interface{}) error {
	return nil
}

func (t testComm) Join(room string) error {
	return nil
}

func (t testComm) Leave(room string) error {
	return nil
}

func (t testComm) BroadcastTo(room, event string, args ...interface{}) error {
	return nil
}

func TestQueuePlayers(t *testing.T) {
	Convey("Players should be added to the queue", t, func() {
		Convey("When they are not already in the queue", func() {
			g := domain.Game{
				Lobby: domain.NewLobby(),
			}
			p := domain.Player{
				Comm: testComm{
					ID: "testID",
				},
			}
			So(QueuePlayers(g, p), ShouldBeTrue)
		})
		Convey("But not when they are already in the queue", func() {
			g := domain.Game{
				Lobby: domain.NewLobby(),
			}
			p := domain.Player{
				Comm: testComm{
					ID: "testID",
				},
			}
			QueuePlayers(g, p)
			So(QueuePlayers(g, p), ShouldBeFalse)
		})
	})
}

//
// func TestGroupPlayers(t *testing.T) {
// 	Convey("Players should be grouped", t, func() {
// 		Convey("When the min number of players is available", func() {
// 			g := domain.Game{
// 				MinPlayers: 2
// 				Lobby: domain.NewLobby(),
// 			}
// 			p := domain.Player{
// 				Comm: testComm{
// 					ID: "testID",
// 				},
// 			}
// 			p2 := domain.Player{
// 				Comm: testComm{
// 					ID: "testID2",
// 				},
// 			}
// 			QueuePlayers(g, p)
// 			QueuePlayers(g, p2)
// 			So(GroupPlayers(g, gi *GamesInfo))
// 			So(QueuePlayers(g, p), ShouldBeTrue)
// 		})
// 		Convey("But not when they are already in the queue", func() {
// 			g := domain.Game{
// 				Lobby: domain.NewLobby(),
// 			}
// 			p := domain.Player{
// 				Comm: testComm{
// 					ID: "testID",
// 				},
// 			}
// 			QueuePlayers(g, p)
// 			So(QueuePlayers(g, p), ShouldBeFalse)
// 		})
// 	})
// }
