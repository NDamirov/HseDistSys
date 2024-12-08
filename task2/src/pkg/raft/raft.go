package raft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"raft/pkg/config"
	"raft/pkg/storage"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type Raft struct {
	sync.Mutex

	metaInfo MetaInfo
	config   *config.Config
	client   *http.Client

	storage storage.Storage

	logs        Log
	syncedIdx   map[int]int
	commitIndex int

	lastHeartbeatTime time.Time
	lastTry           time.Time
	electionTimeout   time.Duration
}

func NewRaft(config *config.Config) *Raft {
	initStatus := Follower
	if config.LeaderOnStart {
		initStatus = Leader
	}
	raft := &Raft{
		metaInfo: MetaInfo{
			Term:     0,
			Status:   initStatus,
			LeaderID: -1,
			VotedFor: -1,
		},
		logs: []LogEntry{
			{
				Base: Base{
					Term: 0,
				},
				Command: OpInit,
			},
		},
		syncedIdx:   make(map[int]int),
		commitIndex: 0,
		config:      config,
		client:      &http.Client{Timeout: config.ResponseTimeout},
		storage:     *storage.NewStorage(),

		lastHeartbeatTime: time.Now(),
		lastTry:           time.Now(),
		electionTimeout:   0,
	}
	for _, port := range config.OtherPorts {
		raft.syncedIdx[port] = 0
	}
	log.Printf("Initialized raft on port %d", config.ServerPort)
	return raft
}

func (r *Raft) Start() error {
	e := echo.New()

	client := e.Group("/api")
	client.POST("/create", r.CreateRequestHandler)
	client.GET("/read", r.ReadRequestHandler)
	client.POST("/update", r.UpdateRequestHandler)
	client.POST("/delete", r.DeleteRequestHandler)
	client.POST("/cas", r.CASRequestHandler)
	client.GET("/get_replicas", r.GetReplicasRequestHandler)

	raft := e.Group("/raft")
	raft.POST("/request_vote", r.RequestVoteRequestHandler)
	raft.POST("/add_log", r.AddLogRequestHandler)

	go r.WaitHeartbeat()
	go r.Heartbeat()

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", r.config.ServerPort),
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return e.StartServer(s)
}

func (r *Raft) RequestVoteRequestHandler(c echo.Context) error {
	var request RequestVoteRequest
	if err := c.Bind(&request); err != nil {
		return err
	}

	r.Lock()
	r.metaInfo.Lock()
	defer r.metaInfo.Unlock()
	defer r.Unlock()

	response := RequestVoteResponse{
		Base: Base{
			Term: r.metaInfo.Term,
		},
		Success: false,
	}

	if request.Term < r.metaInfo.Term {
		return c.JSON(http.StatusOK, response)
	}

	if r.metaInfo.VotedFor == -1 || request.Term > r.metaInfo.Term || r.metaInfo.VotedFor == request.CandidateID {
		r.metaInfo.Term = request.Term
		r.metaInfo.Status = Follower
		r.metaInfo.VotedFor = request.CandidateID

		response.Success = true
	}

	return c.JSON(http.StatusOK, response)
}

func (r *Raft) AddLogRequestHandler(c echo.Context) error {
	var request AppendEntriesRequest
	if err := c.Bind(&request); err != nil {
		return err
	}
	r.metaInfo.Lock()
	r.Lock()
	defer r.metaInfo.Unlock()
	defer r.Unlock()

	log.Printf("Received append request from %d with term %d", request.LeaderID, request.Term)
	if request.Term < r.metaInfo.Term {
		return c.JSON(http.StatusOK, AppendEntriesResponse{
			Base: Base{
				Term: r.metaInfo.Term,
			},
			Success: false,
		})
	}

	r.metaInfo.Term = request.Term
	r.metaInfo.Status = Follower
	r.metaInfo.VotedFor = -1
	r.lastHeartbeatTime = time.Now()

	if request.ParentLogIndex >= len(r.logs) || r.logs[request.ParentLogIndex].Term != request.ParentLogTerm {
		return c.JSON(http.StatusOK, AppendEntriesResponse{
			Base: Base{
				Term: r.metaInfo.Term,
			},
			Success: false,
		})
	}

	r.logs = r.logs[:request.ParentLogIndex+1]
	r.logs = append(r.logs, request.Entries...)
	for i := r.commitIndex + 1; i <= request.LeaderCommitIndex; i++ {
		r.Apply(r.logs[i])
	}
	r.commitIndex = request.LeaderCommitIndex

	return c.JSON(http.StatusOK, AppendEntriesResponse{
		Base: Base{
			Term: r.metaInfo.Term,
		},
		Success: true,
	})
}

func (r *Raft) SendRequestVoteRequest(address string, request RequestVoteRequest) (*RequestVoteResponse, error) {
	log.Printf("Sending request vote request to %s", address)
	request.Term = r.metaInfo.Term

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Post(address+"/raft/request_vote", "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response RequestVoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (r *Raft) SendAppendRequest(address string, request AppendEntriesRequest) (*AppendEntriesResponse, error) {
	log.Printf("Sending append request to %s", address)
	request.Term = r.metaInfo.Term

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Post(address+"/raft/add_log", "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to send append request to %s", address)
	}
	var response AppendEntriesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}
