package main

import (
  "net/rpc"
  "log"
	"net"
  "net/http"
  "time"
  "os"
  "fmt"
  "flag"
  // "strconv"
)

type Nothing struct{}

type Request struct {
	User    string
	Target  string
	Message string
}

type Messages struct {
	Users map[string][]string
}

type handler func(*Messages)
type Server chan<- handler

var (
	DEFAULT_PORT = "8080"
	DEFAULT_HOST = "localhost"
)



func ServerHandler() Server {
  handler := make(chan handler)
  state := &Messages {
		Users: make(map[string][]string),
	}

	go func() {
		for f := range handler { // grab one item at a time from the channel
			f(state) // the actor runs the function that it's handed.
		}
	}()

  return handler
}



func (server Server) Register(username string, reply *Nothing) error {
  finished := make(chan struct{})
  server <- func(elt *Messages) {
    message := username + " logged in"
    for k, _ := range elt.Users {
      elt.Users[k] = append(elt.Users[k], message) // every k user gets the message
    }
    // elt.Users[username] = []string{"Hello from register"}
    elt.Users[username] = nil
    finished <- struct{}{} // the actor feeds struct into the pipe to let ->this func() of completion
  }
  <-finished  // We see a struct coming through pipe so we can now hit return
  return nil
}



func (server Server) CheckMessages(username string, reply *[]string) error {
  finished := make(chan struct{})
  server <- func(elt *Messages) {
    *reply = elt.Users[username]
    elt.Users[username] = nil
    finished <- struct{}{}
  }
  <-finished
  return nil
}



func (server Server) Tell(request Request, reply *Nothing) error {
  finished := make(chan struct{})
  server <- func(elt *Messages) {
    username := request.User
    target := request.Target
    message := request.Message
    // log.Printf("message: %s", message)
    msg := fmt.Sprintf("%s whispers you '%s'", username, message)
    elt.Users[target] = append(elt.Users[target], msg)
    finished <- struct{}{}
  }
  <-finished
  return nil
}


func (server Server) Say(request Request, reply *Nothing) error {
  finished := make(chan struct{})
  server <- func(elt *Messages) {
    msg := request.User + " says " + "'"+ request.Message + "'"
    // log.Printf("message: %s", msg)
    for k, _ := range elt.Users {
      elt.Users[k] = append(elt.Users[k], msg) // every k user gets the message
    }
    finished <- struct{}{}
  }
  <-finished
  return nil
}



func (server Server) List(request Request, reply *Nothing) error {
	finished := make(chan struct{})
	server <- func(elt *Messages) {
		for target, _ := range elt.Users {
			elt.Users[request.User] = append(elt.Users[request.User], target)
		}
		finished <- struct{}{}
	}

	<-finished
	return nil
}


func (server Server) Quit(request Request, reply *Nothing) error {
  finished := make(chan struct{})
  server <- func(elt *Messages) {
    for target, value := range elt.Users {
      if target == request.User {
        delete(elt.Users, target)
      } else {
        message := request.User + " logged out"
        value = append(value, message)
        elt.Users[target] = value
      }
    }
    finished <- struct{}{}
  }
  <-finished
  return nil
}



func (server Server) Help(request Request, reply *Nothing) error {
  finished := make(chan struct{})
  server <- func(elt *Messages) {
    message := "tell <user> some message \n" + "say message \n" + "list \n" + "shutdown"
    elt.Users[request.User] = append(elt.Users[request.User], message)
    finished <- struct{}{}
  }
  <-finished
  return nil
}




func (server Server) Shutdown(username string, reply *Nothing) error {
  log.Print("hello from shutdown called by " + username)
  done := make(chan struct{})      // channel private to the Shutdown method
  server <- func(elt *Messages) { // hand the work off to the actor or server
    time.Sleep(time.Second*5)      // this is the work
    done <- struct{}{}             // more work
  }                                // done with work
  <-done
  os.Exit(0)
  return nil
}



func parseFlags() {
  flag.StringVar(&DEFAULT_PORT, "port", "3410", "port for chat server to listen on")
  flag.Parse()

  DEFAULT_PORT = ":" + DEFAULT_PORT
}



func RunServer(server Server) {
  rpc.Register(server)
  rpc.HandleHTTP()

  l, err := net.Listen("tcp", DEFAULT_PORT)
  if err != nil {
    log.Fatal("listen error:", err)
  }

  log.Printf("Listening on port %s...\n", DEFAULT_PORT)

  if err := http.Serve(l, nil); err != nil {
    log.Fatalf("http.Server: %v", err)
  }
}



func main() {
  server := ServerHandler()
  parseFlags()
  RunServer(server)
}
