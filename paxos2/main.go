package main

import (
  "flag"
  "log"
  "bufio"
  "os"
  "strings"
  "fmt"
)

var (
  gKill     chan bool
  mAddress  string
  gMe  *Replica
)

func propose(cmd string, args []string) {
  var propose Command
  if (len(args) == 2) {
    propose.Command = fmt.Sprintf("%s %s %s", cmd, args[0], args[1])
  } else {
    propose.Command = fmt.Sprintf("%s %s", cmd, args[0])
  }
  log.Printf("propose.Command: [%s] ", propose.Command)
}


func readCommand() (string, []string) {
  reader := bufio.NewReader(os.Stdin)
  command, _ := reader.ReadString('\n')
  command = strings.TrimSpace(command)
  return strings.Fields(command)[0], strings.Fields(command)[1:]
}

func getCommands() {
  for {
    cmd, args := readCommand()
    if cmd == "put" {
      propose(cmd, args[:2])
    }
    log.Printf("cmd: [%v], args: [%v] ", cmd, args[:])
  }
}

func init() {
  gKill = make(chan bool, 1)
  mAddress = getLocalAddress()
}

 func main() {
   flag.Parse()
   mAddress += ":" + flag.Args()[0]
   gMe = createReplica(mAddress, flag.Args()[1:])
   // the other address
   for _,v := range (gMe.Cell) {
     log.Printf("Known Replicas: [%s]", v)
   }

   go getCommands()

   <-gKill
 }
