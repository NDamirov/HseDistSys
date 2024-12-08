package raft

import "sync"

type Status int

const (
	Follower Status = iota
	Candidate
	Leader
)

type MetaInfo struct {
	sync.Mutex
	Term     int
	Status   Status
	LeaderID int
	VotedFor int
}
