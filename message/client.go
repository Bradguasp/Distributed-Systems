package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"net"
	"strconv"
	"errors"
	"flag"
	"net/rpc"
	"time"
)

type Nothing struct{}

type Client struct {
	Username 	string
	Address  	string
	Client   	*rpc.Client
	logged		chan bool
}

type Request struct {
	User    string
	Target  string
	Message string
}



func (c *Client) ClientConnect() (*rpc.Client, error) {
	var err error
	c.Client, err = rpc.DialHTTP("tcp", c.Address)
	if err != nil {
		log.Panicf("Error establishing connection with host: %q", err)
	}
	return c.Client, err
}



func (c *Client) Register() {
	var reply string
	var err error
	c.Client, err = c.ClientConnect()
	if err != nil {
    log.Panicf("Error establishing connection with host: %q", err)
  }
	c.Client.Call("Server.Register", c.Username, &reply)
	log.Printf("%s", reply)
}



func CheckMessages(c *Client) {
	var err error
	c.Client, err = c.ClientConnect()
	if err != nil {
		log.Panicf("Error establishing connection with host: %q", err)
	}

	for {
		time.Sleep(100 * time.Millisecond)

		var reply []string

		err = c.Client.Call("Server.CheckMessages", c.Username, &reply)
		if err != nil {
			log.Fatal("GetMessages error:", err)
		}
		if reply != nil {
			for i := 0; i < len(reply); i++ {
				fmt.Println(reply[i])
			}
		}
	}
}



func (c *Client) Tell(request *Request) {
	var err error
	var reply Nothing
	c.Client, err = c.ClientConnect() // get connection
	if err != nil {
		log.Panicf("Error establishing connection with host: %q", err)
	}
	if err = c.Client.Call("Server.Tell", request, &reply); err != nil {
		log.Fatalf("client.Call: %v", err)
	}
}



func (c *Client) Say(request *Request) {
	var err error
	var reply Nothing
	c.Client, err = c.ClientConnect() // get connection
	if err != nil {
		log.Panicf("Error establishing connection with host: %q", err)
	}
	if err = c.Client.Call("Server.Say", request, &reply); err != nil {
		log.Fatalf("client.Call: %v", err)
	}
}

func (c *Client) List(request *Request) {
	var err error
	var reply Nothing
	c.Client, err = c.ClientConnect() // get connection
	if err != nil {
		log.Panicf("Error establishing connection with host: %q", err)
	}
	if err = c.Client.Call("Server.List", request, &reply); err != nil {
		log.Fatalf("client.Call: %v", err)
	}
}



func (c *Client) Quit(request *Request) {
	var err error
	var reply Nothing
	c.Client, err = c.ClientConnect() // get connection
	if err != nil {
		log.Panicf("Error establishing connection with host: %q", err)
	}
	if err = c.Client.Call("Server.Quit", request, &reply); err != nil {
		log.Fatalf("client.Call: %v", err)
	}
}



func (c *Client) Help(request *Request) {
	var err error
	var reply Nothing
	c.Client, err = c.ClientConnect() // get connection
	if err != nil {
		log.Panicf("Error establishing connection with host: %q", err)
	}
	if err = c.Client.Call("Server.Help", request, &reply); err != nil {
		log.Fatalf("client.Call: %v", err)
	}
}



func (c *Client) Shutdown() {
  var junk Nothing
  var err error
  c.Client, err = c.ClientConnect() // get connection
  if err != nil {
    log.Panicf("Error establishing connection with host: %q", err)
  }
  if err = c.Client.Call("Server.Shutdown", "a 'sudo' user", &junk); err != nil {
    log.Fatalf("client.Call: %v", err)
  }
}


var (
	DEFAULT_PORT = 8080
	DEFAULT_HOST = "localhost"
)
func createClient() (*Client, error) {
	var c *Client = &Client{}
	var host string
	flag.StringVar(&c.Username, "user", "Default_User", "Your username")
	flag.StringVar(&host, "host", "localhost", "The host you want to connect to")
	flag.Parse()

	if !flag.Parsed() {
		return c, errors.New("Unable to create user from commandline flags. Please try again")
	}

	if len(host) != 0 {
		if strings.HasPrefix(host, ":") {
			c.Address = DEFAULT_HOST + host
		} else if strings.Contains(host, ":") {
			c.Address = host
		} else {
			c.Address = net.JoinHostPort(host, strconv.Itoa(DEFAULT_PORT))
		}

	} else {
		c.Address = net.JoinHostPort(DEFAULT_HOST, strconv.Itoa(DEFAULT_PORT)) // Default
	}

	return c, nil
}



func mainLoop(c *Client) {
	for {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			args := string(scanner.Text())
			splitArgs := strings.Split(args, " ")
			var tellMsg string
			var sayMsg string

			if len(splitArgs) > 2 {
				tellMsg = strings.Join(splitArgs[2:], " ")
			}
			if len(splitArgs) >= 1 && strings.HasPrefix(args, "say") {
				sayMsg = strings.Join(splitArgs[1:], " ")
			}

			switch splitArgs[0] {
			case "shutdown":
				c.Shutdown()
			case "tell":
				var request Request
				request.User = c.Username
				request.Target = splitArgs[1]
				request.Message = tellMsg
				c.Tell(&request)
			case "say":
				var request Request
				request.User = c.Username
				request.Message = sayMsg
				c.Say(&request)
			case "list":
				var request Request
				request.User = c.Username
				c.List(&request)
			case "quit":
				var request Request
				request.User = c.Username
				c.Quit(&request)
				fmt.Println("Logging out... ")
				os.Exit(0)
			case "help":
				var request Request
				request.User = c.Username
				c.Help(&request)
			}
		}
  }
}



func main() {
  myClient, err := createClient()
  if err != nil {
    log.Panicf("Error creating client: %q", err)
  }

	myClient.logged = make(chan bool, 1)

	myClient.Register()
	go mainLoop(myClient)
	go CheckMessages(myClient)

	<-myClient.logged
}
