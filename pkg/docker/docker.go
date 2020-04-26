package docker

import (
	"errors"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var (
	reqPsLineSplit = regexp.MustCompile(`[\s]{2,}`)
)

type Container struct {
	ID string
	//Image   string
	//Command string
	Running bool
	//Ports   string
	Names string
}

type Docker struct {
	mu         sync.Mutex
	containers map[string]Container
}

func (d *Docker) Toggle() error {
	var cmd *exec.Cmd
	var action string

	if d.IsRunning() {
		action = "stop"
	} else {
		action = "start"
	}

	if path, err := exec.LookPath("systemctl"); err == nil { // systemd
		cmd = exec.Command(path, action, "docker")
	} else if path, err := exec.LookPath("service"); err == nil { // service
		cmd = exec.Command(path, "docker", action)
	} else if path, err := exec.LookPath("rc-service"); err == nil { // rc-update
		cmd = exec.Command(path, "docker", action)
	} else {
		return errors.New("the initialization system is not defined")
	}

	log.Println(cmd.String())

	return cmd.Run()
}

func (d *Docker) ToggleContainer(containerID string) error {
	container, ok := d.containers[containerID]
	if !ok {
		return errors.New("container is not exist")
	}

	var action string
	if container.Running {
		action = "stop"
	} else {
		action = "start"
	}

	cmd := exec.Command("docker", action, containerID)

	log.Println(cmd.String())

	return cmd.Run()
}

func (d *Docker) IsRunning() bool {
	_, err := d.Containers(true)
	return err == nil
}

func (d *Docker) Containers(force bool) (map[string]Container, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !force && d.containers != nil {
		return d.containers, nil
	}

	cmd := exec.Command("docker", "ps", "-a")
	b, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	d.containers = make(map[string]Container)

	list := strings.Split(string(b), "\n")

	for i := 1; i < len(list); i++ {
		params := reqPsLineSplit.Split(list[i], -1)
		if len(params) > 5 {
			container := Container{}
			container.ID = params[0]
			//container.Image = params[1]
			//container.Command = params[2]
			container.Running = strings.HasPrefix(params[4], "Up")
			if len(params) > 6 {
				//container.Ports = params[5]
				container.Names = params[6]
			} else {
				container.Names = params[5]
			}
			d.containers[container.ID] = container
		}
	}

	return d.containers, nil
}
