package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/exp/rand"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Name       string
	ServerPort int
	OtherPorts []int

	LeaderHeartbeatDuration  time.Duration
	FollowerHeartbeatWaiting time.Duration
	ResponseTimeout          time.Duration

	voteDurationMin int
	voteDurationMax int

	LeaderOnStart bool
}

type yamlConfig struct {
	Ports []int `yaml:"ports"`

	VoteDuration struct {
		Min int `yaml:"min"`
		Max int `yaml:"max"`
	} `yaml:"vote_duration"`

	Timeout struct {
		Leader struct {
			Heartbeat int `yaml:"heartbeat"`
		} `yaml:"leader"`
		Follower struct {
			LeaderHeartbeat int `yaml:"leader_heartbeat"`
		} `yaml:"follower"`
		Response int `yaml:"response"`
	} `yaml:"timeout"`
}

func NewConfig(hostsPath string) (*Config, error) {
	yamlData, err := os.ReadFile(hostsPath)
	if err != nil {
		return nil, err
	}

	var yc yamlConfig
	if err := yaml.Unmarshal(yamlData, &yc); err != nil {
		return nil, err
	}

	portStr := os.Getenv("RAFT_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}
	name := os.Getenv("RAFT_NAME")

	otherPorts := make([]int, 0)
	for _, p := range yc.Ports {
		if p != port {
			otherPorts = append(otherPorts, p)
		}
	}

	if len(otherPorts) == len(yc.Ports) {
		return nil, fmt.Errorf("server port %d not found in ports", port)
	}

	leaderOnStart := false
	if _, ok := os.LookupEnv("RAFT_LEADER_ON_START"); ok {
		leaderOnStart = true
	}

	rand.Seed(uint64(time.Now().UnixNano()))
	return &Config{
		Name:                     name,
		ServerPort:               port,
		OtherPorts:               otherPorts,
		voteDurationMin:          yc.VoteDuration.Min,
		voteDurationMax:          yc.VoteDuration.Max,
		LeaderHeartbeatDuration:  time.Duration(yc.Timeout.Leader.Heartbeat) * time.Millisecond,
		FollowerHeartbeatWaiting: time.Duration(yc.Timeout.Follower.LeaderHeartbeat) * time.Millisecond,
		ResponseTimeout:          time.Duration(yc.Timeout.Response) * time.Millisecond,
		LeaderOnStart:            leaderOnStart,
	}, nil
}

func (c *Config) GetVoteDuration() time.Duration {
	duration := rand.Intn(c.voteDurationMax-c.voteDurationMin+1) + c.voteDurationMin
	return time.Duration(duration) * time.Millisecond
}
