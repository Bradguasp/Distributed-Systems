package main

// Proposers
// Acceptors
// Learners or deciders?

import (
  "sync"
  "fmt"
)

type Nothing struct{}

type Replica struct {
  Address string
  Cell []string
  Slot []Slot // represents a single command to apply to database
  Data map[string]string
  ToApply int
  Mutex sync.Mutex
  Listeners map[string]chan string
  PrepareReplies []chan PrepareResponse
}

func (my Sequence) Cmp(your Sequence) int {
  if my.Number < your.Number {
    return -1
  } else if (my.Number > your.Number) {
    return 1
  } else {
    return 0
  }
}

func (r Replica) String() string {
  return fmt.Sprintf("%s | %s ", r.Address, r.Cell)
}

type PrepareRequest struct {
  Slot Slot
  Sequence Sequence
  SlotNumber int
}

type PrepareResponse struct {
  Okay bool
  Promised Sequence
  Command Command
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
  LargestN int
}

type AcceptRequest struct {
  Slot Slot
  Sequence Sequence
  Command Command
  SlotNumber int
}

type AcceptResponse struct {
  Okay bool
  Promised Sequence
}
