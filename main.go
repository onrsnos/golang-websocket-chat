package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rsms/gotalk"
)

// type Socket struct {
// 	Name     string `json:"name"`
// 	messages []*Message
// }

type Socket struct {
	Name        string `json:"name"`
	SocketToken *gotalk.WebSocket
}
type Message struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}

type Name struct {
	Name string `json:"name"`
}
type Messages struct {
	messages []*Message
}

type Names struct {
	names []*Name
}
type SocketData struct {
	sockets []*Socket
}

func (message *Messages) appendMessageSocket(m *Message) {

	messagesdata.messages = append(messagesdata.messages, m)

	for s := range messagesdata.messages {
		fmt.Println(messagesdata.messages[s].Author)
		fmt.Println(messagesdata.messages[s].Body)
	}
	fmt.Println(messagesdata)
}

type NewMessage struct {
	Message Message `json:"message"`
}
type SocketMap map[string]*Socket
type MessagesType []Messages

var (
	socks        map[*gotalk.WebSocket]int
	socksmu      sync.RWMutex
	messagesdata Messages
	socketdata   SocketData
	namesdata    Names
)

func onConnect(s *gotalk.WebSocket) {

	socksmu.Lock()
	defer socksmu.Unlock()
	socks[s] = 1
	s.CloseHandler = func(s *gotalk.WebSocket, _ int) {
		for i := range socketdata.sockets {
			if socketdata.sockets[i].SocketToken == s {
				fmt.Println("tokenlar eşleşti")
				if i < len(socketdata.sockets) {
					fmt.Println("socket lengti token indexinden büyük")
					socketdata.sockets[i] = socketdata.sockets[len(socketdata.sockets)-1]
					socketdata.sockets[len(socketdata.sockets)-1] = nil
					socketdata.sockets = socketdata.sockets[:len(socketdata.sockets)-1]

					namesdata.names[i] = namesdata.names[len(namesdata.names)-1]
					namesdata.names[len(namesdata.names)-1] = nil
					namesdata.names = namesdata.names[:len(namesdata.names)-1]
				} else {
					fmt.Println("socket lengti token indexine eşit")
					socketdata.sockets[i] = nil
					socketdata.sockets = socketdata.sockets[:i-1]

					namesdata.names[i] = nil
					namesdata.names = namesdata.names[:i-1]
				}

				break
				fmt.Println("Socket data sockets intern")
				fmt.Println(socketdata.sockets)
				fmt.Println("Socket data sockets intern")

			}
		}

		fmt.Printf("Token: %s bağlantısı kesildi!\n", s)
		delete(socks, s)
	}

	fmt.Printf("Token: %s ile  %s adresinden bağlantı sağlanıldı\n", s, s.Conn().LocalAddr())

	s.Notify("showmessages", messagesdata.messages)

	// isimlerden birini random çek
	username := randomName()
	namesdata.names = append(namesdata.names, &Name{username})
	socketdata.sockets = append(socketdata.sockets, &Socket{username, s})
	fmt.Println("Socket data sockets onconnect")
	fmt.Println(socketdata.sockets)
	fmt.Println("Socket data sockets onconnect")
	// fmt.Println("Socket data:")
	// fmt.Println(socketdata.sockets[0].SocketToken)

	for i := range socketdata.sockets {
		fmt.Println(i)

		socketdata.sockets[i].SocketToken.Notify("denemenotification", namesdata.names)
	}

	// fmt.Println(username)
	// soketdeki UserData kolonuna bu ismi ata
	s.UserData = username
	s.Notify("username", username)
}

func broadcast(name string, in interface{}) {
	// socksmu.RLock()
	// defer socksmu.RUnlock()
	for s := range socks {
		s.Notify(name, in)
	}
}

var names struct {
	First []string
	Last  []string
}

func randomName() string {
	first := names.First[rand.Intn(len(names.First))]
	return first
}

func main() {
	socks = make(map[*gotalk.WebSocket]int)

	if namesjson, err := ioutil.ReadFile("names.json"); err != nil {
		panic("failed to read names.json: " + err.Error())
	} else if err := json.Unmarshal(namesjson, &names); err != nil {
		panic("failed to read names.json: " + err.Error())
	}
	rand.Seed(time.Now().UTC().UnixNano())

	gotalk.Handle("send-message", func(s *gotalk.Sock, r NewMessage) error {
		if len(r.Message.Body) == 0 {
			// hata üretildi. hata içeriği  = mesaj boş olamaz
			return errors.New("mesaj boş olamaz")
		}
		fmt.Printf("mesaj geldi. gelen mesaj: %s\n", r.Message.Body)
		username, _ := s.UserData.(string)
		messagesdata.appendMessageSocket(
			&Message{username, r.Message.Body})

		r.Message.Author = username
		broadcast("newmsg", r.Message)
		return nil
	})

	gh := gotalk.WebSocketHandler()
	gh.OnConnect = onConnect
	routes := &http.ServeMux{}
	server := &http.Server{Addr: "localhost:8080", Handler: routes}
	routes.Handle("/gotalk/", gh)
	routes.Handle("/", http.FileServer(http.Dir(".")))

	done := enableGracefulShutdown(server, 5*time.Second)
	fmt.Printf("Listening on http://%s/\n", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
	<-done
}

func enableGracefulShutdown(server *http.Server, timeout time.Duration) chan struct{} {
	server.RegisterOnShutdown(func() {
		fmt.Printf("bağlantı koparılıyor\n")

		socksmu.RLock()
		defer socksmu.RUnlock()
		for s := range socks {
			s.CloseHandler = nil // avoid deadlock on socksmu (also not needed)
			s.Close()
		}
	})
	done := make(chan struct{})
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)
	go func() {
		<-quit // wait for signal

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		fmt.Printf("bağlantı koparılacak\n")

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			fmt.Printf("Server kapatılırken %s  HATASI OLUŞTU\n", err)
		}

		fmt.Printf("bağlantı koparıldı\n")
		close(done)
	}()
	return done
}
