package raft

import (
	"log"
	"sync"
	"time"
)

func (r *Raft) WaitHeartbeat() {
	for {
		time.Sleep(time.Millisecond * 100)
		r.Lock()
		r.metaInfo.Lock()

		if r.metaInfo.Status == Candidate && time.Since(r.lastTry) < r.electionTimeout {
			r.metaInfo.Unlock()
			r.Unlock()
			continue
		}
		if r.metaInfo.Status == Candidate {
			r.BecomeCandidate()
			r.lastTry = time.Now()
			r.electionTimeout = r.config.GetVoteDuration()
			r.metaInfo.Unlock()
			r.Unlock()
			continue
		}

		if r.metaInfo.Status != Follower || time.Since(r.lastHeartbeatTime) < r.config.FollowerHeartbeatWaiting {
			if r.metaInfo.Status != Follower {
				r.lastHeartbeatTime = time.Now()
			}
			r.Unlock()
			r.metaInfo.Unlock()
			continue
		}

		log.Printf("Starting election")

		r.BecomeCandidate()
		if r.metaInfo.Status == Candidate {
			r.lastTry = time.Now()
			r.electionTimeout = r.config.GetVoteDuration()
			log.Printf("Election timeout: %s", r.electionTimeout)
		}

		r.metaInfo.Unlock()
		r.Unlock()
	}
}

func (r *Raft) BecomeCandidate() {
	r.metaInfo.Status = Candidate
	r.metaInfo.VotedFor = r.config.ServerPort
	r.metaInfo.Term++

	req := RequestVoteRequest{
		Base: Base{
			Term: r.metaInfo.Term,
		},
		CandidateID:  r.config.ServerPort,
		LastLogIndex: len(r.logs) - 1,
		LastLogTerm:  r.logs[len(r.logs)-1].Term,
	}

	results := []RequestVoteResponse{}
	mtx := sync.Mutex{}
	wg := sync.WaitGroup{}
	for _, port := range r.config.OtherPorts {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			res, err := r.SendRequestVoteRequest(GetAddress(port), req)
			if err != nil {
				return
			}
			mtx.Lock()
			results = append(results, *res)
			mtx.Unlock()
		}(port)
	}
	wg.Wait()

	acked := 1
	for _, res := range results {
		if res.Success {
			acked++
		}
		if res.Term > r.metaInfo.Term {
			r.metaInfo.Status = Follower
			r.metaInfo.Term = res.Term
			return
		}
	}

	if acked*2 > len(r.config.OtherPorts) {
		r.metaInfo.Status = Leader
	}
}
