package main

import (
  "flag"
  "log"
  "bufio"
  "os"
  "strings"
  "fmt"
  "math/rand"
)

var (
  gKill     chan bool
  mAddress  string
  mReplica  *Replica
)

func propose(cmd string, args []string) {
  var proposal Command
  if (len(args) == 2) {
    proposal.Command = fmt.Sprintf("%s %s %s", cmd, args[0], args[1])
  } else {
    proposal.Command = fmt.Sprintf("%s %s", cmd, args[0])
  }
  log.Printf("proposal.Command: [%s] ", proposal.Command)
  proposal.Address = mAddress
  proposal.Tag = rand.Int()
  var reply bool
  call(mAddress, "Replica.Propose", proposal, &reply)
  // log.Printf("Replica.Propose reply: [%v] ", reply)

  // continue here
  if (reply) {
    var junk Nothing
    // var reply string // reply to fill up. replace old variable
    if cmd == "put" {
      for _, c := range(append(mReplica.Cell, mReplica.Address)) {
        var elt KeyValue
        elt.Key = args[0]
        elt.Value = args[1]
        call(c, "Replica.Put", elt, &junk)
      }
    }
  }
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
    var message string
    for key, value := range mReplica.Database {
      message += fmt.Sprintf("Key: [%s] Value: [%s]\n", key, value)
    }
    log.Printf(message)
    // log.Printf("database [%v] ", mReplica.Database)
  }
}

func init() {
  gKill = make(chan bool, 1)
  mAddress = getLocalAddress()
}

 func main() {
   flag.Parse()
   mAddress += ":" + flag.Args()[0]
   mReplica = createReplica(mAddress, flag.Args()[1:])
   // the other address
   for _,v := range (mReplica.Cell) {
     log.Printf("Known Replicas: [%s]", v)
   }

   go getCommands()

   <-gKill
 }
