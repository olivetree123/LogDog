package utils

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/minio/minio/pkg/wildcard"
	. "logDog/common"
)

type Container struct {
	ID   string
	Name string
}

func ContainerList(patterns []string) ([]Container, error) {
	containers, err := DockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, nil
	}
	var cts []Container
	for _, container := range containers {
		name := container.Names[0][1:]
		flag := false
		for _, pattern := range patterns {
			status := wildcard.MatchSimple(pattern, name)
			if !status {
				Logger.Info("Not match")
				continue
			}
			flag = true
		}
		if !flag {
			continue
		}
		c := Container{
			ID:   container.ID,
			Name: name,
		}
		cts = append(cts, c)
	}
	return cts, nil
}
