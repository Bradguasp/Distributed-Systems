package main

import (
  "sync"
)

type PrepareReply struct {
	Okay     bool
	Promised Sequence
	Command  Command
	Slot Slot
}

type Sequence struct {
  Number  int
  Address string
}

type Command struct {
  Command string
  Address string
  Tag     int
}

type Slot struct {
  Decide    bool
  Command   Command

  Promise   Sequence
  Accepted  Command
}

type Replica struct {
  Database        map[string]string
  Address         string
  Cell            []string
  Slot            []Slot
  Mutex           sync.Mutex
  ToApply         int
  Listeners       map[string]chan string
  PrepareReplies  []chan PrepareReply
}
