package main

import (
  "os"
  "bufio"
  "log"
  "flag"
  "fmt" // doesnt print time and date
  "strings"
  "math/rand"
)

var (
  gShutdown chan bool
  mAddress string
  mReplica *Replica
)

func propose(cmd string, args []string) {
  var proposal Command
  if (len(args) == 2) {
    proposal.Command = fmt.Sprintf("%s %s %s", cmd, args[0], args[1])
  } else {
    proposal.Command = fmt.Sprintf("%s %s", cmd, args[0])
  }
  proposal.Address = mAddress
  proposal.Tag = rand.Int()
  fmt.Printf("cmd[%s] / %s | tag[%v]", proposal.Command, proposal.Address, proposal.Tag)
  var reply bool
  call(mAddress, "Replica.Propose", proposal, &reply)
  // more work

}

func dump() {
  var junk Nothing
  var reply Nothing
  call(mAddress, "Replica.Dump", junk, &reply)
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
    } else if (cmd == "dump") {
      dump()
    }
  }
}

func init() {
  gShutdown = make(chan bool, 1)
  mAddress = getLocalAddress()
}

func main() {
  flag.Parse()
  log.Printf("local address is [%s] ", mAddress)
  mAddress += ":" + flag.Args()[0]
  mReplica = createReplica(mAddress, flag.Args()[1:])

  fmt.Printf("\nCell | %v\n", mReplica)

  go getCommands()

  <- gShutdown
}
