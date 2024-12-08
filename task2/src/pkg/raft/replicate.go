package raft

import (
	"errors"

	"github.com/labstack/gommon/log"
)

func (r *Raft) Replicate(entry LogEntry) error {
	r.Lock()
	defer r.Unlock()

	entry.Term = r.metaInfo.Term
	r.logs = append(r.logs, entry)
	results := make(chan bool, len(r.config.OtherPorts))

	for _, port := range r.config.OtherPorts {
		go func(port int) {
			req := AppendEntriesRequest{
				Entries:           []LogEntry{entry},
				ParentLogIndex:    len(r.logs) - 2,
				ParentLogTerm:     r.logs[len(r.logs)-2].Term,
				LeaderCommitIndex: r.commitIndex,
				LeaderID:          r.config.ServerPort,
			}

			succ, err := r.SendAppendRequest(GetAddress(port), req)
			if err != nil {
				results <- false
				return
			}
			results <- succ.Success
		}(port)
	}

	acked := 0
	for i := 0; i < len(r.config.OtherPorts); i++ {
		if <-results {
			acked++
		}
	}
	if acked*2 > len(r.config.OtherPorts) {
		for i := r.commitIndex + 1; i < len(r.logs); i++ {
			r.Apply(r.logs[i])
		}
		for key := range r.syncedIdx {
			r.syncedIdx[key]++
		}
		r.commitIndex = len(r.logs) - 1
		return nil
	}
	r.logs = r.logs[:len(r.logs)-1]
	return errors.New("failed to replicate")
}

func (r *Raft) Apply(entry LogEntry) error {
	switch entry.Command {
	case OpCreate:
		return r.storage.Create(entry.Key, *entry.Value)
	case OpSet:
		return r.storage.Set(entry.Key, *entry.Value)
	case OpCAS:
		return r.storage.CAS(entry.Key, *entry.CompareValue, *entry.Value)
	case OpDelete:
		return r.storage.Delete(entry.Key)
	default:
		log.Warnf("Got strange command number: %d", entry.Command)
	}
	return nil
}
