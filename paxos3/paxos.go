package main

import (

)

// PrepareRequest: Slot, Sequence // PrepareResponse: Okay, Promised, Command,  -> Slot <- this may be useless
func (r *Replica) Prepare(elt PrepareRequest, reply *PrepareResponse) {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  // your number is bigger? Okay ill promise to not accept any number small than that
  if (elt.Sequence.Number > r.ID) {
    reply.Okay = true
    reply.Promised = elt.Slot.Promise
    reply.Command = elt.Slot.Command
    r.ID++
    // your number is small? Sorry I already promised on a number bigger than that
  } else {
    reply.Okay = false
    reply.Promised = r.Slot[r.ID].Promise
    reply.Command = r.Slot[elt.Sequence.Number].Command
    reply.Slot = r.Slot[elt.Sequence.Number] // this line probably doesnt need to be here
  }
  return nil
}

func (r *Replica) Accept(elt AcceptRequest, reply *AcceptResponse) error {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  if (elt.Sequence.Number <= r.ID) {
    // Only if the replica has not promised a value greater than elt.Sequence.Number
    reply.Okay = true
  } else {
    // if there is a value greater than elt.Sequence.Number then False
    // A Higher value than elt.Sequence.Number has been promised for Slot
    reply.Okay = false
    // Promised should be the last value promised for this slot
    reply.Promised = r.ID
  }
  return nil
}
