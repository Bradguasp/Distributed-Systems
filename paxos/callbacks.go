package main

import (
	"fmt"
	"log"
	"strings"
)

func (r *Replica) Prepare(m PrepareSend, reply *PrepareReply) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	//log.Println("replica.Prepare[", r.Address, "] Recieved:", m.String())
	if (m.Sequence.Number > r.ToApply) {
		reply.Okay = true
		reply.Promised = m.Slot.Promise
		reply.Command = m.Slot.Command
		// store promise
		r.ToApply++
		//r.Slot[r.ToApply] = m.Slot
	} else {
		reply.Okay = false
		reply.Promised = r.Slot[r.ToApply].Promise
		reply.Command = r.Slot[m.Sequence.Number].Command  // reply with the command you have stored at that slot
		reply.Slot = r.Slot[m.Sequence.Number]
	}
	return nil
}

func (r *Replica) Accept(m AcceptSend, reply *AcceptReply) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()
	//log.Println("replica.Accept[", r.Address, "] Recieved:", m.String())
	if (m.Sequence.Number <= r.ToApply) {
		reply.Okay = true
	} else {
		reply.Okay = false
		reply.Promised = r.ToApply
	}
	return nil
}

func (r *Replica) Decide(m DecideSend, reply *DecideReply) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	//log.Println("replica.Decide[", r.Address, "] Recieved:", m.String())

	if (m.Slot.Decided == true) {
		r.ToApply = m.Slot.Promise.Number
		//fmt.Println(r.ToApply)
		r.Slot[r.ToApply] = m.Slot
		reply.Okay = true
	} else {
		log.Println("replica.Decide: Error: Asked to apply an undecided slot:", m.Slot.String())
		reply.Okay = false
	}

	return nil
	/*
		if ch, present := elt.Listeners[key]; present {
		  // for each value in map, if value was present ->
		}
	*/
}

func (r *Replica) Dump(junk *Nothing, reply *Nothing) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	fmt.Println("Replica[", r.Address, "]")
	// cell data
	fmt.Print("Cell [")
	for _, c := range r.Cell {
		fmt.Print(c, " ")
	}
	fmt.Println("]")
	// slot data
	fmt.Println("Slots:")
	for i, s := range r.Slot {
		fmt.Printf("[%d] %s\n", i, s.Command.Command)
	}
	// actual data
	fmt.Println("Data:")
	for k, v := range r.Data {
		fmt.Printf("[%s] %s\n", k, v)
	}
	return nil
}

func (r *Replica) Propose(command Command, reply *bool) error{
	r.Mutex.Lock()

	//og.Println("replica.Propose[", r.Address, "]")
	//log.Println(command.String())

	// prepare init
	//r.ToApply++
	var seq Sequence
	seq.Address = gAddress
	seq.Number = r.ToApply+1

	var slot Slot
	slot.Decided = false
	slot.Command = command
	slot.Promise = seq
	slot.Accepted = command

	var ps PrepareSend
	ps.Sequence = seq
	ps.Slot = slot

	var pr PrepareReply
	r.Mutex.Unlock()

	// prepare round
	done := false
	for {
		pp := 0 // positive prepare replies
		np := 0 // negative prepare replies
		//fmt.Println(ps.String())
		for _, c := range (append(r.Cell, r.Address)) {
			call(c, "Replica.Prepare", ps, &pr)
			//fmt.Println("Replica.Prepare returned:", pr.String())
			if (pr.Okay == true) {
				// tally votes
				pp++
				if (pp * 2 >= len(r.Cell) + 1) {
					done = true
					fmt.Println("[", ps.Sequence.Number, "] Prepare accepted:", pp, "/", len(r.Cell) + 1)
					break
				}
			} else if (pr.Okay == false) {
				// why was the prepare rejected
				if (pr.Promised.Number >= ps.Sequence.Number) {
					fmt.Println("[", ps.Sequence.Number, "] Prepare rejected number", ps.Sequence.Number, ". It is not higher than promised number (", pr.Promised.Number, ")")
					fmt.Println("[", ps.Sequence.Number, "] Prepare suggested command:", pr.Command.String())
				}
				// tally votes
				np++
				if (np * 2 >= len(r.Cell) + 1) {
					// SHORTCUT
					// store learned slot and apply it
					r.ToApply++
					r.Slot[r.ToApply] = pr.Slot

					var junk Nothing
					command, args := strings.Fields(pr.Slot.Command.Command)[0], strings.Fields(pr.Slot.Command.Command)[1:]
					if command == "put" {
						var kv KeyValue
						kv.Key = args[0]
						kv.Value = args[1]
						call(r.Address, "Replica.Put", kv, &junk)
					} else if command == "get" {
						var reply2 string
						call(r.Address, "Replica.Get", args[0], &reply2)
						//fmt.Printf("%s has a value of %s\n", args[0], reply2)
					} else if command == "delete" {
						call(r.Address, "Replica.Delete", args[0], &junk)
					}

					// SHORTCUT

					// update n and try again
					fmt.Println("[", ps.Sequence.Number, "] Prepare rejected:", np, "/", len(r.Cell) + 1)
					ps.Sequence.Number++
					ps.Slot.Promise = ps.Sequence
					break
				}
			}
		}
		// prepare has been accepted
		if (done) {
			break
		}
	}

	// accept init
	r.Mutex.Lock()
	var as AcceptSend
	as.Command = pr.Command
	as.Sequence = ps.Sequence

	var ar AcceptReply
	r.Mutex.Unlock()

	// accept round
	pa := 0 // positive accept replis
	na := 0 // negative accept replies
	decided := false
	for _, c := range(append(r.Cell, r.Address)) {
		call(c, "Replica.Accept", as, &ar)
		//fmt.Println("Replica.Accept returned:", ar.String())
		if (ar.Okay == true) {
			pa++
			if (pa*2 >= len(r.Cell)+1) {
				fmt.Println("[", as.Sequence.Number,"] Accept accepted:", pa ,"/", len(r.Cell)+1)
				*reply = true
				decided = true
				break
			}
		} else if (ar.Okay == false) {
			na++
			r.ToApply = ar.Promised // sync with other cells
			if (na*2 >= len(r.Cell)+1) {
				fmt.Println("[", as.Sequence.Number,"] Accept rejected:", na ,"/", len(r.Cell)+1)
				break
			}
		}
	}

	// decide round
	if (decided ==  true){
		// command accepted. send out decide messages
		r.Mutex.Lock()
		slot.Decided = true
		slot.Promise = pr.Promised
		var ds DecideSend
		ds.Command = pr.Command
		ds.Slot = slot
		var dr DecideReply
		r.Mutex.Unlock()

		da := 0 // decide accept
		na := 0 // negative decide
		for _, c := range (append(r.Cell, r.Address)) {
			call(c, "Replica.Decide", ds, &dr)
			if (dr.Okay == true) {
				na++
				if (na == len(r.Cell)) {
					fmt.Println("[", as.Sequence.Number, "] Decide accepted", na, "/", len(r.Cell) + 1)
				}
			} else if (dr.Okay == false) {
				da++
				if (na == len(r.Cell)) {
					fmt.Println("[", as.Sequence.Number, "] Decide rejected", da, "/", len(r.Cell) + 1)
				}
			}
		}

	}

	return nil
}

func (r *Replica) Put(kv KeyValue, reply *Nothing) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	//log.Println("Replica.Put called with KeyValue::", kv.String())
	r.Data[kv.Key] = kv.Value
	return nil
}

func (r *Replica) Get(k string, reply *string) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	//log.Println("Replica.Get called with key:", k)
	//fmt.Println("Replica.Get: ", r.Data[k])
	*reply = r.Data[k]
	return nil
}

func (r *Replica) Delete(k string, reply *Nothing) error {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	//log.Println("Replica.Delete Called with key:", k)
	delete(r.Data, k)
	return nil
}
