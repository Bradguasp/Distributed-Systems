package main

import (
  "flag"
  "log"
  "fmt"
)

var (
  gShutdown chan bool
  mAddress string
  mReplica *Replica
)

func init() {
  gShutdown = make(chan bool, 1)
  mAddress = getLocalAddress()
}

func main() {
  flag.Parse()
  log.Printf("local address is [%s] ", mAddress)
  mAddress += ":" + flag.Args()[0]
  mReplica = runReplica(mAddress, flag.Args()[1:])

  fmt.Printf("\nCell | %v\n", mReplica)

  <- gShutdown
}
