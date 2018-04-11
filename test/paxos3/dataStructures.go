package main

// Proposers
// Acceptors
// Learners or deciders?

import (
  "sync"
  "fmt"
)

type Replica struct {
  Address string
  Cell []string
  Data map[string]string
  Slot []Slot
  ID int
  Mutex sync.Mutex
  Listeners map[string]chan string
  PrepareReplies []chan PrepareReply
}

func (r Replica) String() string {
  return fmt.Sprintf("%s | %s ", r.Address, r.Cell)
}

type PrepareRequest struct {
  Slot Slot
  Sequence Sequence
}

type PrepareResponse struct {
  Okay bool
  Promised Sequence
  Command Command
  Slot Slot
}

type Sequence struct {
  Number int
  Address string
}

type Command struct {
  Command string
  Address string
  Tag int
}

// Each slot represents a single command that should be applied to the database
// the goal of the system is to ensure that every replica agrees on the same operation in each slot
// and applies them to the database in the same order.
type Slot struct {
  Decide bool
  Command Command
  Promise Sequence
  Accepted Command
}

type AcceptRequest struct {
  Slot Slot
  Sequence Sequence
  Command Command
}

type AcceptResponse struct {
  Okay bool
  Promised int
}
