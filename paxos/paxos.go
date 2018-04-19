package main

import (
  "fmt"
  //"log"
  "time"
  "strings"
)

func (r *Replica) Find(yourProposal PrepareRequest, myReply *FindSlot) error {
  r.Mutex.Lock()
  myReply.SlotNumber = r.ToApply
  r.Mutex.Unlock()
  return nil
}

// PrepareRequest: Slot, Sequence // PrepareResponse: Okay, Promised, Command,  -> Slot <- this may be useless
func (r *Replica) Prepare(yourProposal PrepareRequest, myReply *PrepareResponse) error {
  r.Mutex.Lock()
  fmt.Printf("[%v] Prepare: called with N=[%v]\n", r.ToApply, yourProposal.Sequence.Number)
  fmt.Printf("Prepare: [%v][%v]\n", yourProposal.Sequence.Number, r.Slot[yourProposal.SlotNumber].Promise.Number)
  if yourProposal.Sequence.Number > r.Slot[yourProposal.SlotNumber].Promise.Number {
    r.Slot[yourProposal.SlotNumber].Promise.Number = yourProposal.Sequence.Number
    r.Slot[yourProposal.SlotNumber].Promise.Address = yourProposal.Sequence.Address
    // if (yourProposal.Sequence.Number > r.Slot[yourProposal.SlotNumber].) {
    //   r.Slot[yourProposal.SlotNumber].HighestN = yourProposal.Sequence.Number
    // }
    myReply.Okay = true
    myReply.Promised = yourProposal.Slot.Promise
    myReply.Command = r.Slot[yourProposal.SlotNumber].Command
    fmt.Printf("[%v] Prepare: --> Promising N=[%v]/[%v] with no command\n", r.ToApply, myReply.Promised.Number, myReply.Promised.Address)
  } else if (r.Slot[yourProposal.SlotNumber].Promise.Number >= yourProposal.Sequence.Number) {
    myReply.Okay = false
    myReply.Promised = r.Slot[yourProposal.SlotNumber].Promise
    //fmt.Printf("TO APPLY IS: %v\n", r.ToApply)
    //myReply.HighestSlot = r.ToApply
    fmt.Printf("[%v] Prepare: --> Rejecting N=[%v] / because [%v]/[%v] already promised\n", r.ToApply, yourProposal.Sequence.Number, r.Slot[yourProposal.SlotNumber].Promise.Number, r.Slot[yourProposal.SlotNumber].Promise.Address)
  }
  r.Mutex.Unlock()
  return nil
}

// Slot, Sequence, Command  -> Okay Promised
func (r *Replica) Accept(yourProposal AcceptRequest, myReply *AcceptResponse) error {
  r.Mutex.Lock()
  // Accept: called with N=1/:8001 Command={"put r t"}
  fmt.Printf("[%v] Accept: called with N=[%v] Command={%v}\n", r.ToApply, yourProposal.Sequence.Number, yourProposal.Command)
  fmt.Printf("Accept: [%v][%v]\n", yourProposal.Sequence.Number, r.Slot[yourProposal.SlotNumber].Promise.Number)
  if yourProposal.Sequence.Number >= r.Slot[yourProposal.SlotNumber].Promise.Number {
    myReply.Okay = true
    myReply.Promised = r.Slot[yourProposal.SlotNumber].Promise
    r.Slot[yourProposal.SlotNumber].Command = yourProposal.Command
    fmt.Printf("[%v] Accept: --> Accepting N=[%v] / with Command={%v}\n", r.ToApply, yourProposal.Sequence.Number, yourProposal.Command)
  } else if (r.Slot[yourProposal.SlotNumber].Promise.Number > yourProposal.Sequence.Number) {
    myReply.Okay = false
    myReply.Promised = r.Slot[yourProposal.SlotNumber].Promise
    fmt.Printf("[%v] Accept: --> \n", r.ToApply)
  }
  r.Mutex.Unlock()
  return nil
}

// Slot, Command
func (r *Replica) Decide(yourProposal DecideRequest, myReply *DecideResponse) error {
  r.Mutex.Lock()
  r.Mutex.Unlock()
  fmt.Printf("[%v] Decide: called with Command={%v}\n", r.ToApply, yourProposal.Command)
  if (yourProposal.Slot.Decided == true) {
    r.Slot[yourProposal.SlotNumber] = yourProposal.Slot
    myReply.Okay = true
    r.ToApply++
    fmt.Printf("[%v] Decide: --> Applying Command={%v}\n", r.ToApply, yourProposal.Command)

    if err := call(r.Address, "Replica.ApplyCommand", yourProposal.Command, &myReply); err != nil {
    }
  } else {
    myReply.Okay = false
  }
  return nil
}

func (r *Replica) ApplyCommand(command Command, myReply *DecideResponse) error {
  // r.Mutex.Lock()
  // defer r.Mutex.Unlock()
  fmt.Printf("[%v] Apply: called with Command={%v}\n", r.ToApply, command.Command)
  //fmt.Printf("cmd.Command is of type [%T]\n", command.Command)
  if strings.Fields(command.Command)[0] == "put" {
    key := strings.Fields(command.Command)[1]
    value := strings.Fields(command.Command)[2:]
    r.Data[key] = value
  } else if (strings.Fields(command.Command)[0] == "get") {
    iso := strings.Fields(command.Command)[1]
    for key,value := range(r.Data) {
      if key == iso {
        myReply.Key = iso
        myReply.Value = value
        myReply.Address = r.Address
        myReply.Requested = "get"
        return nil
      }
      // fmt.Printf("key %v | value %v\n", key, value)
    }
  } else if (strings.Fields(command.Command)[0] == "delete") {
    toDelete := strings.Fields(command.Command)[1]
    for key,value := range(r.Data) {
      if key == toDelete {
        myReply.Key = toDelete
        myReply.Value = value
        myReply.Address = r.Address
        myReply.Requested = "delete"
        delete(r.Data, key)
        return nil
      }
    }
  }
  return nil
}

func (r *Replica) Propose(cmd Command, reply *bool) error {
  time.Sleep(time.Second*2)
  r.Mutex.Lock()
  // wait := 2
  var seq Sequence
  seq.Address = mAddress
  seq.Number = 1

  var slot Slot
  slot.Decided = false
  slot.Command = cmd
  slot.Promise = seq
  slot.Accepted = cmd
  slot.HighestN = -1

  // package I prepare with this slot and new Sequence to apply a command at this slot number
  var prepare_request PrepareRequest
  prepare_request.Slot = slot
  prepare_request.Sequence = seq
  prepare_request.SlotNumber = r.ToApply
  var prepare_response PrepareResponse

  var find_slot FindSlot
  r.Mutex.Unlock()

  for _, c := range(r.Cell) {
    if err := call(c, "Replica.Find", prepare_request, &find_slot); err != nil {
        fmt.Printf("Error calling function Replica.FindSlot %v", err)
      }
    //fmt.Printf("Going to put it in slot: %v\n", find_slot.SlotNumber)
    //prepare_request.SlotNumber = prepare_response.HighestSlot
    r.ToApply = find_slot.SlotNumber
  }


// PREPARE ROUND
  done := false
  for { // if prepare round fails the first time then find the next highest Sequence Number
    fmt.Printf("[%v] Propose: starting prepare round with N=[%v]\n", r.ToApply, prepare_request.Sequence.Number)
    time.Sleep(time.Second*2)
    upVote := 0
    downVote := 0
    for _, c := range(r.Cell) {
      time.Sleep(time.Second*2)
      //high := 0
      if err := call(c, "Replica.Prepare", prepare_request, &prepare_response); err != nil {
        fmt.Printf("Error calling function Replica.Prepare %v", err)
      }
      if (prepare_response.Okay == true) {
        upVote++
        // if (prepare_response.Promised.Number > prepare_request.Sequence.Number) {
          // fmt.Println("hello")
          // prepare_request.Sequence.Number = prepare_response.Promised.Number
        // }
        fmt.Printf("[%v] Propose: --> yes vote received with no accepted command\n", r.ToApply)
        if (upVote * 2 >= len(r.Cell) + 1) {
          done = true
          // fmt.Println("[", prepare_request)
          fmt.Printf("[%v] Propose: --> got a majority of yes votes, ignoring any additional responses\n", r.ToApply)
          break
        }
      } else if (prepare_response.Okay == false) {
        downVote++
        if (prepare_response.Promised.Number >= prepare_request.Sequence.Number) {
          prepare_request.Sequence.Number = prepare_response.Promised.Number + 1
          fmt.Printf("new n=[%v]\n", prepare_request.Sequence.Number)
          break
          // fmt.Println("hello")
          // prepare_request.Sequence.Number = prepare_response.Promised.Number
        }
        fmt.Printf("[%v] Propose: --> no vote received with Promise=[%v]/[%v]\n", r.ToApply, prepare_response.Promised.Number, prepare_response.Promised.Address)
        if (downVote * 2 >= len(r.Cell) + 1) {
          fmt.Printf("[%v] Propose: --> got a majority of no votes, ignoring any additional responses\n", r.ToApply)
          fmt.Printf("[%v] Propose: --> prepare failed, sleeping\n", r.ToApply)
          time.Sleep(time.Second*2)
          break
        }
      }
    }
    if (done) {
      break
    }
  }
  slot.HighestN = prepare_request.Sequence.Number
  fmt.Printf("[%v] Propose: starting accept round with N=[%v] Command={%v}\n", r.ToApply, slot.HighestN, slot.Command)
// ACCEPT ROUND
  r.Mutex.Lock()
  var accept_request AcceptRequest
  accept_request.Slot = prepare_request.Slot
  accept_request.Sequence = prepare_request.Sequence
  accept_request.Command = cmd
  accept_request.SlotNumber = r.ToApply
  var accept_response AcceptResponse
  // fmt.Printf("slot.command: %v\n", yourProposal.Slot.Command)
  r.Mutex.Unlock()
  upVote := 0
  downVote := 0
  decided := false
  for _, c := range(r.Cell) {
    call(c, "Replica.Accept", accept_request, &accept_response)
    if (accept_response.Okay == true) {
      upVote++
      fmt.Printf("[%v] Propose: -->" + "yes " + "vote received\n", r.ToApply)
      if (upVote * 2 >= len(r.Cell) + 1) {
        decided = true
        fmt.Printf("[%v] Propose: --> got a majority of " + "yes " + "votes, ignoring any additional responses\n", r.ToApply)
        break
      }
    } else if (accept_response.Okay == false) {
      downVote++
      fmt.Printf("---accept_response = %v \n", accept_response.Promised.Number)
      if (downVote * 2 >= len(r.Cell) + 1) {
        fmt.Printf("[%v] Propose: --> got a majority of no votes, ignoring any additional responses\n", r.ToApply)
        time.Sleep(time.Second*2)
        break
      }
    }
  }

  fmt.Printf("[%v] Propose: starting decide round with Command={%v}\n", r.ToApply, slot.Command)
// DECIDE ROUND

  if (decided == true) {
    r.Mutex.Lock()
    slot.Decided = true
    slot.Promise = prepare_response.Promised
    var decide_request DecideRequest
    decide_request.Slot = slot
    decide_request.Command = cmd
    decide_request.SlotNumber = r.ToApply
    var decide_response DecideResponse
    r.Mutex.Unlock()

    upVote := 0
    downVote := 0
    for _, c := range(r.Cell) {
      call(c, "Replica.Decide", decide_request, &decide_response)
      if (decide_response.Okay == true) {
        upVote++
        //fmt.Println("[", accept_request.SlotNumber, "] Decide accpeted", upVote, "/", len(r.Cell) + 1)
        if decide_response.Requested == "get" {
          fmt.Printf("replica[%v] got [%v]=> found %v\n", decide_response.Address, decide_response.Key, decide_response.Value)
        } else if (decide_response.Requested == "delete") {
          fmt.Printf("replica[%v] got [%v]=> deleted %v\n", decide_response.Address, decide_response.Key, decide_response.Value)
        }
      } else if (decide_response.Okay == false) {
        downVote++
        //fmt.Println("[", accept_request.SlotNumber, "] Decide declined", downVote, "/", len(r.Cell) + 1)
      }
    }
    // need to think more | line below prints empty since i never called myself because of majority
    fmt.Printf("Slot[%v] got command %v\n", r.ToApply, r.Slot[accept_request.SlotNumber].Command)
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
    fmt.Printf("%s %s\n", key, value)
  }
  return nil
}