package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/PavelVershinin/docker_systray_menu/pkg/docker"

	"github.com/getlantern/systray"
)

var (
	dockerCmd       = &docker.Docker{}
	dockerItem      *systray.MenuItem
	containersItems = make(map[string]*systray.MenuItem)
	exitItem        *systray.MenuItem
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon("assets/icon.png"))
	systray.SetTitle("Docker containers")
	systray.SetTooltip("Docker containers")

	dockerItem = systray.AddMenuItem("Docker", "")

	systray.AddSeparator()

	updateMenu()

	systray.AddSeparator()
	exitItem = systray.AddMenuItem("Exit", "")

	go func() {
		for {
			select {
			case <-dockerItem.ClickedCh:
				if err := dockerCmd.Toggle(); err != nil {
					log.Println(err)
				}
			case <-exitItem.ClickedCh:
				os.Exit(0)
			}
		}
	}()

	go func() {
		for {
			<-time.After(time.Millisecond * 500)
			updateMenu()
		}
	}()
}

func onExit() {

}

func getIcon(s string) []byte {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}

func updateMenu() {
	if !dockerCmd.IsRunning() {
		dockerItem.SetTitle("Start docker")
		for _, item := range containersItems {
			item.Disable()
		}
		return
	}

	dockerItem.SetTitle("Stop docker")

	containers, err := dockerCmd.Containers(true)
	if err != nil {
		log.Println(err)
		return
	}

	for _, container := range containers {
		if _, ok := containersItems[container.ID]; !ok {
			containersItems[container.ID] = systray.AddMenuItem("", "")
			go func(containerID string) {
				for range containersItems[containerID].ClickedCh {
					if err := dockerCmd.ToggleContainer(containerID); err != nil {
						log.Println(err)
					}
				}
			}(container.ID)
		} else if containersItems[container.ID].Disabled() {
			containersItems[container.ID].Enable()
		}

		if container.Running {
			containersItems[container.ID].SetTitle("Stop " + container.Names)
		} else {
			containersItems[container.ID].SetTitle("Start " + container.Names)
		}
	}
}
