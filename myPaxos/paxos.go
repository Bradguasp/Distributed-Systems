package main

import (
  "fmt"
)

// PrepareRequest: Slot, Sequence // PrepareResponse: Okay, Promised, Command,  -> Slot <- this may be useless
func (r *Replica) Prepare(yourProposal PrepareRequest, myReply *PrepareResponse) error {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  // if yourProposal Sequence > r Sequence
  // my promise on that slot number is less than your number
  if r.Slot[yourProposal.SlotNumber].Promise.Cmp(yourProposal.Sequence) < 0 {
    myReply.Okay = true
    myReply.Promised = yourProposal.Sequence
    myReply.Command = r.Slot[yourProposal.SlotNumber].Command
    r.Slot[yourProposal.SlotNumber].LargestN = yourProposal.Slot.LargestN
  // my number is greater // here is the number that I promised on
  } else {
    myReply.Okay = false
    myReply.Promised = r.Slot[yourProposal.SlotNumber].Promise
  }
  return nil
}

func (r *Replica) Accept(yourProposal AcceptRequest, myReply *AcceptResponse) error {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()

  // mine is less than yours | this < that
  if r.Slot[yourProposal.SlotNumber].Promise.Cmp(yourProposal.Sequence) < 0 {
    myReply.Okay = true
    myReply.Promised = r.Slot[yourProposal.SlotNumber].Promise
    r.Slot[yourProposal.SlotNumber].Command = yourProposal.Command
  // mine is greater than yours | this > that | higher value than Seq has been promised
  } else {
    myReply.Okay = false
    myReply.Promised = r.Slot[yourProposal.SlotNumber].Promise
  }
  return nil
}

func (r *Replica) Propose(cmd Command, reply *bool) error {
  *reply = true
  // more work
  return nil
}

func (r *Replica) Dump(junk *Nothing, reply *Nothing) error {
  fmt.Print("made it to server dump")
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  fmt.Println("Replica [", r.Address, "]")
  fmt.Println("Cell [")
  for _,c := range r.Cell {
    fmt.Print(c, " ")
  }
  fmt.Println("]")
  fmt.Println("Slots: ")
  for key, value := range r.Slot {
    fmt.Printf("[%d] %s\n", key, value.Command.Command)
  }
  fmt.Println("Data: ")
  for key, value := range r.Data {
    fmt.Printf("[%s] %s\n", key, value)
  }
  return nil
}
