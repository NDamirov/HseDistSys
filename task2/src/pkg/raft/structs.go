package raft

type OpCode int

const (
	OpInit OpCode = iota
	OpCreate
	OpGet
	OpSet
	OpCAS
	OpDelete
)

type Base struct {
	Term int `json:"term"`
}

type LogEntry struct {
	Base
	Command      OpCode  `json:"command"`
	Key          string  `json:"key"`
	Value        *string `json:"value"`
	CompareValue *string `json:"compare_value"`
}

type Log = []LogEntry

type AppendEntriesRequest struct {
	Base
	Entries           Log `json:"entries"`
	ParentLogIndex    int `json:"parent_log_index"`
	ParentLogTerm     int `json:"parent_log_term"`
	LeaderCommitIndex int `json:"leader_commit_index"`
	LeaderID          int `json:"leader_id"`
}

type AppendEntriesResponse struct {
	Base
	Success bool `json:"success"`
}

type RequestVoteRequest struct {
	Base
	CandidateID  int `json:"candidate_id"`
	LastLogIndex int `json:"last_log_index"`
	LastLogTerm  int `json:"last_log_term"`
}

type RequestVoteResponse struct {
	Base
	Success bool `json:"success"`
}
