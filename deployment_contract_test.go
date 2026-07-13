package main

import (
	"os"
	"strings"
	"testing"
)

func TestProductionDeploymentUsesIsolatedComposeProject(t *testing.T) {
	restartScript, err := os.ReadFile("restart.sh")
	if err != nil {
		t.Fatalf("read restart.sh: %v", err)
	}

	script := string(restartScript)
	if !strings.Contains(script, `PROJECT_NAME="${COMPOSE_PROJECT_NAME:-swarm-api}"`) {
		t.Fatal("restart.sh must default to an isolated swarm-api Compose project")
	}
	if strings.Contains(script, "COMPOSE_PROJECT_NAME=gratheon") {
		t.Fatal("restart.sh must not join the shared gratheon Compose project")
	}
}
