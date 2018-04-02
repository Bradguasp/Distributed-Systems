package main

import (
  // "log"
  "fmt"
  "strings"
)

func (r *Replica) Prepare(message PrepareSend, reply *PrepareReply) error {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()

  if(message.Sequence.Number > r.ToApply) {
    reply.Okay = true
    reply.Promised = message.Slot.Promise
    reply.Command = message.Slot.Command
    r.ToApply++
  } else {
    reply.Okay = false
    reply.Promised = r.Slot[r.ToApply].Promise
    reply.Command = r.Slot[message.Sequence.Number].Command
    reply.Slot = r.Slot[message.Sequence.Number]
  }
  return nil
}


func (r *Replica) Propose(cmd Command, reply *bool) error {
  r.Mutex.Lock()
  var s Sequence
  s.Number = r.ToApply + 1
  s.Address = mAddress

  var slot Slot
  slot.Decide = false
  slot.Command = cmd
  slot.Promise = s
  slot.Accepted = cmd

  var prepare_send PrepareSend
  prepare_send.Slot = slot
  prepare_send.Sequence = s

  var prepare_reply PrepareReply
  r.Mutex.Unlock()

  // Prepare(slot, seq) ROUND
  done := false
  for {
    upvote := 0
    downvote := 0
    for _,c := range(append(r.Cell, r.Address)) {
      call(c, "Replica.Prepare", prepare_send, &prepare_reply)
      if(prepare_reply.Okay == true) {
        upvote++
        if(upvote * 2 >= len(r.Cell) + 1) {
          done = true
          fmt.Println("[", prepare_send.Sequence.Number, "] Prepare accepted:", upvote, "/", len(r.Cell) + 1)
					break
        }
      } else if (prepare_reply.Okay == false) {
        if(prepare_reply.Promised.Number >= prepare_send.Sequence.Number) {
          fmt.Println("[", prepare_send.Sequence.Number, "] Prepare rejected number", prepare_send.Sequence.Number, ". It is not higher than promised number (", prepare_reply.Promised.Number, ")")
					fmt.Println("[", prepare_send.Sequence.Number, "] Prepare suggested command:", prepare_reply.Command)
        }
        downvote++
        if (downvote * 2 >= len(r.Cell) + 1) {
          r.ToApply++
          r.Slot[r.ToApply] = prepare_reply.Slot

          var junk Nothing
          cmd, args := strings.Fields(prepare_reply.Slot.Command.Command)[0], strings.Fields(prepare_reply.Slot.Command.Command)[1:]
          if cmd == "put" {
            var elt KeyValue
            elt.Key = args[0]
            elt.Value = args[1]
            call(r.Address, "Replica.Put", elt, &junk)
          }
          fmt.Println("[", prepare_send.Sequence.Number, "] Prepare rejected:", downvote, "/", len(r.Cell) + 1)
					prepare_send.Sequence.Number++
					prepare_send.Slot.Promise = prepare_send.Sequence
					break
        }
      }
    }
    if (done) {
      break
    }
  }
  return nil
}

func (r *Replica) Put(elt KeyValue, reply *Nothing) error {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  r.Database[elt.Key] = elt.Value
  return nil
}
