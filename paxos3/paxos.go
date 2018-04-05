package main

import (

)

func (r *Replica) Accept(elt AcceptSend, reply *AcceptReply) error {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  if (elt.Sequence.Number <= r.ToApply) {
    // Only if the replica has not promised a value greater than elt.Sequence.Number
    reply.Okay = true
  } else {
    // A Higher value than elt.Sequence.Number has been promised for Slot
    reply.Okay = false
    // Promised should be the last value promised for this slot
    reply.Promised = r.ToApply
  }
  return nil
}
