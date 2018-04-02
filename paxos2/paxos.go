package main

func (r *Replica) Propose(cmd Command, reply *bool) error {
  *reply = true
  return nil
}

func (r *Replica) Put(elt KeyValue, reply *Nothing) error {
  r.Mutex.Lock()
  defer r.Mutex.Unlock()
  r.Database[elt.Key] = elt.Value
  return nil
}
