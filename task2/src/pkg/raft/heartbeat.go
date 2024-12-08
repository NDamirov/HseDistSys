package raft

import (
	"log"
	"sync"
	"time"
)

func (r *Raft) Heartbeat() {
	for {
		time.Sleep(r.config.LeaderHeartbeatDuration)

		r.Lock()
		r.metaInfo.Lock()

		if r.metaInfo.Status != Leader {
			r.metaInfo.Unlock()
			r.Unlock()
			continue
		}
		log.Printf("Sending heartbeat to %d followers", len(r.config.OtherPorts))

		wg := sync.WaitGroup{}
		toFollower := 0
		toFollowerMtx := sync.Mutex{}

		for _, port := range r.config.OtherPorts {
			wg.Add(1)
			go func(port int) {
				defer wg.Done()
				last := r.syncedIdx[port]
				req := AppendEntriesRequest{
					LeaderID:          r.config.ServerPort,
					LeaderCommitIndex: r.commitIndex,
					ParentLogIndex:    last,
					ParentLogTerm:     r.logs[last].Term,
					Entries:           r.logs[last+1:],
				}
				res, err := r.SendAppendRequest(GetAddress(port), req)
				if err != nil {
					log.Printf("Error sending heartbeat to %d: %v", port, err)
					return
				}
				if res.Success {
					return
				}
				if res.Term > r.metaInfo.Term {
					toFollowerMtx.Lock()
					toFollower++
					toFollowerMtx.Unlock()
				} else {
					r.syncedIdx[port]--
				}
			}(port)
		}
		wg.Wait()

		if toFollower > 0 {
			r.metaInfo.Status = Follower
		}

		r.metaInfo.Unlock()
		r.Unlock()
	}
}
