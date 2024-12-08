package main

import (
	"fmt"
	"log"
	"raft/pkg/config"
	"raft/pkg/raft"
)

func main() {
	config, err := config.NewConfig("/app/config/server.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}

	log.SetFlags(log.Ltime | log.Lshortfile)
	log.SetPrefix(fmt.Sprintf("[RAFT] [%s] ", config.Name))

	raft := raft.NewRaft(config)
	if raft == nil {
		log.Fatalln("Failed to create raft")
	}
	err = raft.Start()
	log.Fatalf("Server failed: %s", err)
}
