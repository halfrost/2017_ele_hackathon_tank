// Package consts defines some common use constants.
package consts

const (
	// BufferSize represents the size of buffered transport.
	BufferSize int = 4096
	// EnvDockerContainerID is an environment variable which indicates the docker container id.
	EnvDockerContainerID string = "MESOS_TASK_ID"
)
