package main

import (
  "fmt"
  "log"
  "time"
)

// PrepareRequest: Slot, Sequence // PrepareResponse: Okay, Promised, Command,  -> Slot <- this may be useless
func (r *Replica) Prepare(yourProposal PrepareRequest, myReply *PrepareResponse) error {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  // Your number is > mine
  // if r.Slot[yourProposal.SlotNumber].Promise.Cmp(yourProposal.Sequence) < 0 {
  if yourProposal.Sequence.Number > r.Slot[yourProposal.Slot.SlotNumber].Promise.Number {
    myReply.Okay = true
    myReply.Promised = yourProposal.Slot.Promise
    fmt.Println("Promised: ", myReply.Promised)
    myReply.Command = r.Slot[yourProposal.Slot.SlotNumber].Command
    // fmt.Printf("the command is: %s", r.Slot[yourProposal.SlotNumber].Command)
    r.Slot[yourProposal.Slot.SlotNumber].Promise = yourProposal.Slot.Promise
  // } else if (r.Slot[yourProposal.SlotNumber].Promise.Cmp(yourProposal.Sequence) > 0) {
    // My nuber is >= yours
  } else if (r.Slot[yourProposal.Slot.SlotNumber].Promise.Number >= yourProposal.Sequence.Number) {
    log.Println("prepare greater than")
    myReply.Okay = false
    myReply.Promised = r.Slot[yourProposal.Slot.SlotNumber].Promise
  }
  return nil
}

// Slot, Sequence, Command  -> Okay Promised
func (r *Replica) Accept(yourProposal AcceptRequest, myReply *AcceptResponse) error {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  // log.Printf("%v Was called", r.Address)
  // mine is less than yours | this < that
  if r.Slot[yourProposal.Slot.SlotNumber].Promise.Cmp(yourProposal.Sequence) < 0 {
    // log.Println("accept less than")
    myReply.Okay = true
    myReply.Promised = r.Slot[yourProposal.Slot.SlotNumber].Promise
    r.Slot[yourProposal.Slot.SlotNumber].Command = yourProposal.Command
  // mine is greater than yours | this > that | higher value than Seq has been promised
  } else {
    // log.Println("accept greater than")
    myReply.Okay = false
    myReply.Promised = r.Slot[yourProposal.Slot.SlotNumber].Promise
  }
  return nil
}

// Slot, Command
func (r *Replica) Decide(yourProposal DecideRequest, myReply *DecideResponse) error {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()

  if (yourProposal.Slot.Decided == true) {
    r.Slot[yourProposal.Slot.SlotNumber] = yourProposal.Slot
    myReply.Okay = true
    r.ToApply++
  } else {
    myReply.Okay = false
  }
  return nil
}

func (r *Replica) Propose(cmd Command, reply *bool) error {
  time.Sleep(time.Second*2)
  r.Mutex.Lock()

  wait := 2

  var seq Sequence
  seq.Address = mAddress
  seq.Number = 1

  var slot Slot
  slot.Decided = false
  slot.Command = cmd
  slot.Promise = seq
  slot.Accepted = cmd
  slot.SlotNumber = r.ToApply

  // package I prepare with this slot and new Sequence to apply a command at this slot number
  var prepare_request PrepareRequest
  prepare_request.Slot = slot
  prepare_request.Sequence = seq
  // prepare_request.SlotNumber = r.ToApply
  fmt.Printf("Starting prepare round with N=[%v] at slot [%v]\n", prepare_request.Slot.SlotNumber, prepare_request.Slot.SlotNumber)

  var prepare_response PrepareResponse
  r.Mutex.Unlock()
// PREPARE ROUND
  done := false
  for { // if prepare round fails the first time then find the next highest Sequence Number
    upVote := 0
    downVote := 0
    for _, c := range(append(r.Cell, r.Address)) {
      if err := call(c, "Replica.Prepare", prepare_request, &prepare_response); err != nil {
        fmt.Printf("Error calling function Replica.Prepare %v", err)
      }
      fmt.Printf("     --> promising me with N=[%v] \n", prepare_response.Promised.Number)
      if (prepare_response.Okay == true) {
        upVote++
        if (upVote * 2 >= len(r.Cell) + 1) {
          done = true
          // fmt.Println("[", prepare_request)
          break
        }
      } else {
        downVote++
        prepare_request.Slot.Promise.Number = prepare_response.Promised.Number + 1
        fmt.Printf("     --> Rejecting N=[%v] \n", prepare_response.Promised.Number)
        fmt.Printf("   new Sequence N=[%v] \n", prepare_request.Slot.Promise.Number)
        if (downVote * 2 >= len(r.Cell) + 1) {
          time.Sleep(time.Second*2)
          wait += 1
        }
      }
    }
    if (done) {
      break
    }
  }

// ACCEPT ROUND
  r.Mutex.Lock()
  var accept_request AcceptRequest
  accept_request.Slot = slot
  accept_request.Sequence = seq
  accept_request.Command = cmd
  // accept_request.SlotNumber = 0
  var accept_response AcceptResponse
  // fmt.Printf("slot.command: %v\n", yourProposal.Slot.Command)
  r.Mutex.Unlock()
  upVote := 0
  downVote := 0
  decided := false
  for _, c := range(append(r.Cell, r.Address)) {
    call(c, "Replica.Accept", accept_request, &accept_response)
    if (accept_response.Okay == true) {
      upVote++
      fmt.Printf("+++accept_response = %v \n", accept_response.Promised.Number)
      if (upVote * 2 >= len(r.Cell) + 1) {
        decided = true
        break
      }
    } else if (accept_response.Okay == false) {
      downVote++
      fmt.Printf("---accept_response = %v \n", accept_response.Promised.Number)
      if (downVote * 2 >= len(r.Cell) + 1) {

      }
    }
  }

// DECIDE ROUND

  if (decided == true) {
    fmt.Println("majority said yes")
    r.Mutex.Lock()
    slot.Decided = true
    slot.Promise = prepare_response.Promised
    var decide_request DecideRequest
    decide_request.Slot = slot
    decide_request.Command = prepare_response.Command
    var decide_response DecideResponse
    r.Mutex.Unlock()

    upVote := 0
    downVote := 0
    for _, c := range(append(r.Cell, r.Address)) {
      call(c, "Replica.Decide", decide_request, &decide_response)
      if (decide_response.Okay == true) {
        upVote++
        fmt.Println("[", accept_request.Slot.SlotNumber, "] Decide accpeted", upVote, "/", len(r.Cell) + 1)
      } else if (decide_response.Okay == false) {
        downVote++
        fmt.Println("[", accept_request.Slot.SlotNumber, "] Decide declined", downVote, "/", len(r.Cell) + 1)
      }
    }
    // r.Slot[yourProposal.SlotNumber].Command = yourProposal.Command
    // at this point. the other 2 applied the command to their slots
    // need to think more | line below prints empty since i never called myself because of majority
    fmt.Println("Command Stored [", r.Slot[accept_request.Slot.SlotNumber].Command, "]")
  }

  return nil
}

func (r *Replica) Dump(junk *Nothing, reply *Nothing) error {
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
