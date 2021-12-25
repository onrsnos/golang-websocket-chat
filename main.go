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

type Socket struct {
	Name     string `json:"name"`
	messages []*Message
}

type Message struct {
	Author string `json:"author"`
	Body   string `json:"body"`
}

type Messages struct {
	messages []*Message
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
	sockets      SocketMap
	messagesdata Messages
)

func onConnect(s *gotalk.WebSocket) {

	socksmu.Lock()
	defer socksmu.Unlock()
	socks[s] = 1
	s.CloseHandler = func(s *gotalk.WebSocket, _ int) {
		fmt.Printf("Token: %s bağlantısı kesildi!\n", s)
		delete(socks, s)
	}

	fmt.Printf("Token: %s ile  %s adresinden bağlantı sağlanıldı\n", s, s.Conn().LocalAddr())

	s.Notify("showmessages", messagesdata.messages)

	// isimlerden birini random çek
	username := randomName()
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
