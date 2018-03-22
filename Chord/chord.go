package main

import (
	"bufio"
	"crypto/sha1"
	"flag"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strings"
	"strconv"
	"time"
)

func hashString(s Address) *big.Int {

	h := sha1.New()
	h.Write([]byte(string(s)))
	return new(big.Int).SetBytes(h.Sum(nil))
}

const keySize = sha1.Size * 8
const SuccessorsListSize = 3
var two = big.NewInt(2)
var hashMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(keySize), nil)

var force_strings = strings.Split("It looks like Sandpeople did this, all right. Look, here are Gaffi sticks, Bantha tracks. It's just...I never heard of them hitting anything this big before. They didn't. But we are meant to think they did. These tracks are side by side. Sandpeople always ride single file to hide there numbers. These are the same Jawas that sold us Artoo and Threepio. And these blast points, too accurate for Sandpeople. Only Imperial stormtroopers are so precise. Why would Imperial troops want to slaughter Jawas? If they traced the robots here, they may have learned who they sold them to. And that would lead them home! Wait, Luke! It's too dangerous. Uncle Owen! Aunt Beru! Uncle Owen! Hey, you're not permitted in there. It's restricted. You'll be deactivated for sure.. Don't call me a mindless philosopher, you overweight glob of grease! Now come out before somebody sees you. Secret mission? What plans? What are you talking about? I'm not getting in there! I'm going to regret this. There goes another one. Hold your fire. There are no life forms. It must have been short-circuited. That's funny, the damage doesn't look as bad from out here. Are you sure this things safe? Are they away? They have just made the jump into hyperspace. You're sure the homing beacon is secure aboard their ship? I'm taking an awful risk, Vader. This had better work. Hey...hey, open the pressure maintenance hatch on unit number... where are we? Three-two-six-eight-two-seven. If we can just avoid any more female advice, we ought to be able to get out of here. Well, let's get moving! Where are you going? No, wait. They'll hear! Come here, you big coward! Chewie! Come here! Listen. I don't know who you are, or where you came from, but from now on, you do as I tell you. Okay? Look, Your Worshipfulness, let's get one thing straight! I take orders from one person! Me! It's a wonder you're still alive. Will somebody get this big walking carpet out of my way? No reward is worth this. Well, that's the trick, isn't it? And it's going to cost you something extra. Ten thousand in advance. Ten thousand? We could almost buy our own ship for that! But who's going to fly it, kid! You? You bet I could. I'm not such a bad pilot myself! We don't have to sit here and listen... We haven't that much with us. But we could pay you two thousand now, plus fifteen when we reach Alderaan. Seventeen, huh! Okay. You guys got yourself a ship. We'll leave as soon as you're ready. Docking bay Ninety-four. Ninety-four. Looks like somebody's beginning to take an interest in your handiwork. All right, we'll check it out.", " ")

func jump(address Address, fingerentry int) *big.Int {
	n := hashString(address)
	fingerentryminus1 := big.NewInt(int64(fingerentry) -1)
	jump := new(big.Int).Exp(two, fingerentryminus1, nil)
	sum := new(big.Int).Add(n, jump)

	return new(big.Int).Mod(sum, hashMod)
}

func between(start, elt, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}

type Nothing struct {}

type FindArgument struct {
	Found bool
	Address Address
}

type Address string

func (a Address) String() string {
	if a == "" {
		return "(empty address)"
	}
	s := fmt.Sprintf("%040x", hashString(a))
	return s[:8] + "..(" + string(a) + ")"
}

type Key string

func (k Key) String() string {
	if k == "" {
		return "(empty key)"
	}
	s:= fmt.Sprintf("%040x", hashString(Address(k)))
	return s[:8] + "..(" + string(k) + ")"
}

type Node struct {
	successor Address
	me Address
	predecessor Address
	successors []Address
	bucket map[Key]string
	fingers [161]Address
}

func (s Server) Notify(args *Address, reply *Nothing) error {
	finished := make (chan struct{})
	s <- func (n *Node) {
		start := hashString(n.predecessor)
		end := hashString(n.me)
		elt := hashString(*args)
		if n.predecessor == "" || between(start, elt, end, false) {
			n.predecessor = *args
		}
		finished <- struct {}{}
	}
	<-finished
	return nil
}

func (s Server) GetSuccessors(args *Nothing, reply *[]Address) error {
	finished := make(chan struct{})
	s <- func (n *Node) {
		*reply = []Address{n.me}
		// end := int(math.Min(float64(len(n.successors)), SuccessorsListSize - 1))
		r := append(*reply, n.successors...)
		if len(r) >= SuccessorsListSize {
			r = r[:SuccessorsListSize]
		}
		*reply = r
		finished <- struct {}{}
	}
	<- finished
	return nil
}

func (n *Node) Stabilize() {
	client, err := rpc.DialHTTP("tcp", string(n.successor))
	if err != nil {
		log.Printf("We have lost a successor")
		for i := range(n.successors) {
			log.Printf("%s", n.successors[i])
		}
		n.successors = n.successors[1:]
		if len(n.successors) == 0 {
			n.successor = n.me
			n.successors = []Address{n.me}
		} else {
			n.successor = n.successors[0]
		}
		return
	}
	var predecessor Address
	var junk Nothing
	if err := client.Call("Server.GetPredecessor", &junk, &predecessor); err != nil {
		log.Printf("Server.GetPredecessor: %v", err)
		closeClient(client);
		return
	}
	closeClient(client)
	if predecessor != "" {
		if between(hashString(n.me), hashString(predecessor), hashString(n.successor), false) {
			n.successor = predecessor
		}
	}

	// log.Printf("Before stablize getsuccessors %v", predecessor)
	client, err = rpc.DialHTTP("tcp", string(n.successor))
	if err != nil {
		log.Printf("%v, %s : %s", n.successors, n.successor, n.predecessor)
		closeClient(client)
		return
	}
	if err := client.Call("Server.GetSuccessors", &junk, &n.successors); err != nil {
		log.Printf("Server.GetSuccessors: %s", err)
		closeClient(client)
		return
	}
	// log.Printf("after stablize getsuccessors")

	if err := client.Call("Server.Notify", &n.me, &junk); err != nil {
		log.Printf("Server.Notify: %v", err)
		closeClient(client)
		return
	}
	closeClient(client)
	// log.Printf("stablize/ ending")

}



func (n *Node) fix_fingers (me Address) {
	_, err := rpc.DialHTTP("tcp", string(n.successor))
	if err != nil {
		return
	}
	for next := 1; next < 161; next++ {
		// log.Printf("before fix_fingers nextsuccessor")
		nextSuccessor := find(jump(me, next), n.successor)
		n.fingers[next] = nextSuccessor
	}
}

func (s Server) Find_successor(id *big.Int, reply *FindArgument) error {
	finished := make(chan struct {})
	s <- func(n *Node) {
		start := hashString(n.me)
		end := hashString(n.successor)
		if between(start, id, end, true) {
			reply.Found = true
			reply.Address = n.successor
		} else {
			reply.Found = false
			reply.Address = n.Closest_preceding_node(id)
		}
		finished <- struct {}{}
	}
	<-finished
	return nil
}

func (s Server) Put_all(args *map[Key]string, reply *Nothing) error {
	finished := make(chan struct{})
	s <- func(n *Node) {
		for k, v := range(*args) {
			log.Printf("%s: %s", k , v)
			// append keys and values
			n.bucket[k] = v
		}
		finished <- struct{}{}
	}
	<-finished
	return nil
}


func (n *Node) Closest_preceding_node(id *big.Int) Address {
	// for i := 160; i > 0; i-- {
	// 	if between(hashString(n.me), hashString(n.fingers[i]), id, true) {
	// 		return n.fingers[i]
	// 	}
	// }
	return n.successor
}


func find(id *big.Int, start Address) Address {
	findInfo := FindArgument{}
	findInfo.Found = false
	findInfo.Address = start
	i := 0
	maxSteps := 161
	for !(findInfo.Found) && i < maxSteps {
		client := getClient(findInfo.Address)
		if err := client.Call("Server.Find_successor", id, &findInfo); err != nil {
			log.Fatalf("Server.find_successor: %v", err)
		}
		i++
		closeClient(client)
	}
	if findInfo.Found {
		return findInfo.Address
	}
	log.Fatalf("Could not find successor for %s in find function", id)
	return ""
}

type handler func (*Node)
type Server chan<- handler
type BucketArgument struct {
	Key Key
	Value string
}
func (s Server) Ping(args *Nothing, reply *bool) error {
	finished := make(chan struct{})
	s <- func(n* Node) {
		*reply = true
		finished <- struct {} {}
	}
	<-finished
	return nil
}


func (s Server) Put(args *BucketArgument, reply *Nothing) error  {
	finished := make(chan struct {})
	s <- func (n *Node) {
		n.bucket[args.Key] = args.Value
		log.Printf("server put: [%s] => [%s]", args.Key, args.Value)
		finished <- struct{}{}
	}
	<-finished
	return nil
}

func (s Server) Get(args Key, reply *string) error {
	finished := make(chan struct{})
	s <- func (n *Node) {
		log.Printf("Address of owner : %s", n.me)
		*reply = n.bucket[args]
		finished <- struct{} {}
	}
	<-finished
	return nil
}

func (s Server) GetSuccessor(args *Nothing, reply *Address) error {
	finished := make(chan struct{})
	s <- func (n *Node) {
		*reply = n.successor
		finished <- struct {} {}
	}
	<-finished
	return nil
}

func (s Server) GetPredecessor(args *Nothing, reply *Address) error {
	finished := make(chan struct{})
	s <- func (n *Node) {
		*reply = n.predecessor
		finished <- struct {} {}
	}
	<-finished
	return nil
}


func (s Server) Delete(args Key, reply *Nothing) error {
	finished := make(chan struct{})
	s <- func (n *Node) {
		delete(n.bucket, args)
		finished <- struct {}{}
	}
	<-finished
	return nil
}

func (s Server) Dump(args *Nothing, reply *Nothing) error {
	finished := make(chan struct{})
	s <- func (n *Node) {
		log.Printf("---------")
		log.Printf("Neighborhood")
		log.Printf("pred:   " + string(n.predecessor))
		log.Printf("self:   " + string(n.me))
		for s := range(n.successors) {
			log.Printf("succ %s: %s", s, n.successors[s])
		}
		log.Printf("\n")
		log.Printf("Finger Table")
		for i := range(n.fingers) {
			if (i > 0 && i < 160 && n.fingers[i] != n.fingers[i+1]) || i == 160 {
				log.Printf("\t\t[%s]: %s", i, n.fingers[i])
			}
		}
		log.Printf("\n")

		log.Printf("Data items: ")
		for d := range(n.bucket) {
			log.Printf("\t\t[%s] => [%s]", d, n.bucket[d])
		}
		finished <- struct {}{}
	}
	<-finished
	return nil
}

func (s Server) Get_all(args *Address, reply *map[Key]string) error {
	finished := make(chan struct{})
	s <- func (n *Node) {
		start := hashString(n.predecessor)
		if n.predecessor == "" {
			start = hashString(n.me)
		}
		end:= hashString(*args)
		if len(*reply) == 0 {
			*reply = make(map[Key]string)
		}
		log.Printf("Me: %s\nPredecessor: %s\nArgs: %s", n.me, n.predecessor, args)
		for k, v := range(n.bucket) {
			log.Printf("%s : %s", k, v)
			elt := hashString(Address(k))
			// log.Printf("%v :: %v :: %v / %v ", k, elt, end, start)
			// var value bool
			// inclusive := true
			// log.Printf("cmp end & start : %v", end.Cmp(start))
			// if end.Cmp(start) > 0 {
			// 	log.Printf("cmp start & elt : %v", start.Cmp(elt))
			// 	log.Printf("cmp elt & end : %v", elt.Cmp(end))
			// 	value = (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
			// } else {
			// 	value = start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
			// }
			if between(start, elt, end, true) {
			// if value {

				(*reply)[k] = v
				delete(n.bucket, k)
			}
		}
		finished <- struct{}{}
	}
	<-finished
	return nil
}


func startActor(successor Address, me Address) Server {
	log.Printf("startActor Me: %s", me)
	ch := make(chan handler)
	state := new(Node)
	state.bucket = make(map[Key]string)
	if successor != me{
		s := find(hashString(me), successor)
		state.successor = s
		state.successors = []Address{s}
		client := getClient(state.successor)
		if client == nil {
			return nil
		}
		if err := client.Call("Server.Get_all", me, &state.bucket); err != nil {
			log.Fatalf("Server.Get_all: %v", err)
		}
		closeClient(client)
	} else {
		state.successor = successor
		state.successors = []Address{successor}
	}
	state.me = me

	go func(){
		for f:= range ch {
			f(state)
		}
	}()
	go func() {
		for {
			state.Stabilize()
			_, err := rpc.DialHTTP("tcp", string(state.predecessor));
			if err != nil {
				state.predecessor = ""
			}
			state.fix_fingers(me)
			time.Sleep(time.Second)
		}
	}()
	return ch
}



func server (machine, port string, address Address) bool {
	actor := startActor(address, Address(machine+":"+port))
	if actor == nil {
		return false
	}
	rpc.Register(actor)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":" + port)
	if e != nil {
		log.Printf("Listen error: %v", e)
		return false
	}
	go func () {
		if err := http.Serve(l, nil); err != nil {
			log.Printf("http.serve: %v", err)
			return
		}
	}()
	return true
}

func getClient(address Address) *rpc.Client {
	client, err := rpc.DialHTTP("tcp", string(address))
	if err != nil {
		log.Printf("client.dialHTTP: %v", err)
		return nil
	}
	return client
}

func closeClient(c *rpc.Client) {
	if err := c.Close(); err != nil {
		log.Fatalf("client.close: %v", err)
	}
}


func main() {
	var port string
	var machine string
	flag.StringVar(&port, "port", "8000", "Port that will be used to connect to chord.")
	flag.StringVar(&machine, "machine", "localhost", "Machine that the node will be hosted on")
	scanner := bufio.NewScanner(os.Stdin)
	flag.Parse()
	listening := false
	var junk Nothing
	me := Address(machine + ":" + port)
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		switch line[0] {
		case "port":
			if ! listening {
				port = line[1]
				fmt.Println("Port set to " + port)
			} else {
				fmt.Println("cannot change port after joining a ring")
			}
		case "create":
			listening = true
			if server(machine, port, me) == false {
				return
			}
			fmt.Println("Creating a ring on " + port)
		case "join":
			listening = true
			 if server(machine, port, Address(line[1])) == false {
				 return
			 }
			log.Printf("join: You have joined [%s]", line[1])
		case "ping":
			client, err := rpc.DialHTTP("tcp", line[1])
			if  err != nil {
				log.Fatalf("client.dialhtttp: %s", err)
			}
			var reply bool
			if err = client.Call("Server.Ping", &junk, &reply); err != nil {
				log.Fatalf("Server.Ping %v", err)
			}
			log.Printf("[%s] has pinged: [%s]", line[1], reply)
			if err = client.Close(); err != nil {
				log.Fatalf("client.Close: %v", err)
			}
		case "get":
			log.Printf("%s", hashString(Address(line[1])))
			a := find(hashString(Address(line[1])), me)
			client := getClient(a)
			var reply string
			key := Key(line[1])
			if err := client.Call("Server.Get", &key, &reply); err != nil {
				log.Fatalf("Server.Get: %v", err)
			}
			log.Printf("[%s] found [%s]", key, reply)
			closeClient(client)
		case "put":
			a := find(hashString(Address(line[1])), me)
			args := new(BucketArgument)
			args.Key = Key(line[1])
			args.Value = strings.Join(line[2:], " ")

			log.Printf("Looking for [%s] on address [%v]", line[1], a)
			client := getClient(a)
			if err := client.Call("Server.Put", args, &junk); err != nil {
				log.Fatalf("Server.Put: %v", err)
			}
			log.Printf("put: [%v] => [%v]", args.Key, args.Value)
			closeClient(client)
		case "putrandom":
			d, err := strconv.Atoi(line[1])
			if err != nil {
				log.Printf("Second argument in prutramdom is not in integer. Please use in integer")
				break
			}
			for i := 0; i < d; i++ {
				rkey := force_strings[rand.Int() % len(force_strings)]
				rvalue := force_strings[rand.Int() % len(force_strings)]
				a := find(hashString(Address(rkey)), me)
				args := new(BucketArgument)
				args.Key = Key(rkey)
				args.Value = rvalue

				log.Printf("Looking for [%s] on address [%v]", line[1], a)
				client := getClient(a)
				if err := client.Call("Server.Put", args, &junk); err != nil {
					log.Fatalf("Server.Put: %v", err)
				}
				log.Printf("put: [%v] => [%v]", args.Key, args.Value)
				closeClient(client)
			}

		case "delete":
			a := find(hashString(Address(line[1])), me)
			client := getClient(a)
			if err := client.Call("Server.Delete", line[2], &junk); err != nil {
				log.Fatalf("Server.Delete: %v", err)
			}
			log.Printf("Found and deleted [%v]", line[2])
			closeClient(client)
		case "dump":
			client := getClient(me)
			if err := client.Call("Server.Dump", &junk, &junk); err != nil {
				log.Fatalf("Server.Dump: %v", err)
			}
			closeClient(client)
		case "help":
			fmt.Println("Help goes here")
		case "quit":
			log.Printf("%s", me)
			client := getClient(me)
			var junk Nothing
			data := make(map[Key]string)
			if err := client.Call("Server.Get_all", me, &data); err != nil {
				log.Fatalf("Server.Get_all: %v", err)
			}
			var successor Address
			if err := client.Call("Server.GetSuccessor", &junk, &successor); err != nil {
				log.Fatalf("Server.GetSuccessor : %v", err)
			}
			closeClient(client)
			client = getClient(successor)
			if err := client.Call("Server.Put_all", &data, &junk); err != nil {
				log.Fatalf("Server.Put_all: %v", err)
			}
			closeClient(client)
			os.Exit(0)
		case "\n":
			fmt.Println("")
		default:
			fmt.Println(line[0] + " is not a valid command!")
		}
	}
}
