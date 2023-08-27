package main

import (
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/fsnotify/fsnotify"
)

type ServerConfig struct {
	// Build System Executable
	BuildSystem string `json:"buildSystem"`
	// Build Args
	BuildArgs []string `json:"buildArgs"`

	// Main executable
	Executable string `json:"executable"`
	// Executable args
	Args []string `json:"args"`

	// Folders to watch
	Folders []string `json:"folders"`
}

func main() {
	config := ServerConfig{}
	config_filename := "obsconfig.json"
	initialise, to_run := false, false
	initialised := false
	// Command-line flags
	flag.BoolVar(&initialise, "init", false, "initialise the directory for observation")
	flag.BoolVar(&to_run, "run", false, "run the executable and observe")
	flag.Parse()

	// Reading config file
	config_file, err := os.ReadFile(config_filename)
	if err != nil {
		if initialise {
			dump, err := json.Marshal(config)
			if err != nil {
				log.Fatal(err.Error())
			}
			err = os.WriteFile(config_filename, dump, fs.ModePerm)
			if err != nil {
				log.Fatal(err.Error())
			}
		} else {
			log.Println(err.Error())
			log.Fatalln("May be run \"observe --init\"")
		}

	} else {
		err = json.Unmarshal(config_file, &config)
		if err != nil {
			log.Fatalln(err.Error())
		}

		if len(config.BuildSystem) == 0 || len(config.Executable) == 0 {
			log.Fatalln("No build or run command")
		}
		initialised = true
	}

	if initialised && to_run {
		pipe := make(chan fsnotify.Event)
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		err = watcher.Add(".")
		if err != nil {
			log.Fatal(err)
		}

		for _, v := range config.Folders {
			err = watcher.Add(v)
			if err != nil {
				log.Fatal(err)
			}
		}

		log.Println("Observing...")
		ExecuteCommand(&config)
		go EnQueue(watcher, 100, pipe)
		go ProcessQueue(pipe, &config)
		<-make(chan struct{})
	}
}

// this function fires when there is any kind of file changes in the watched folder
// and adds to the local queue if it is not repeated
func EnQueue(watcher *fsnotify.Watcher, delay int, pipe chan fsnotify.Event) {
	millis := time.Millisecond * time.Duration(delay)
	var lastEvent fsnotify.Event

	timer := time.NewTimer(millis)
	timeout := false

	for {
		select {
		case lastEvent = <-watcher.Events:
			if timeout {
				pipe <- lastEvent
				timeout = false
			}
		case err := <-watcher.Errors:
			log.Println(err)
		case <-timer.C:
			timeout = true
			timer.Reset(millis)
		}
	}
}

// This function fires when there is an event in the local queue
func ProcessQueue(pipe chan fsnotify.Event, config *ServerConfig) {
	var event fsnotify.Event
	for {
		if event.Has(fsnotify.Write) {
			log.Println("Restarting...")
			ExecuteCommand(config)
		}
		event = <-pipe
	}
}

func ExecuteCommand(config *ServerConfig) {

	build_command := exec.Command(config.BuildSystem, config.BuildArgs...)
	build_command.Stdin = os.Stdin
	build_command.Stdout = os.Stdout
	build_command.Stderr = os.Stderr

	commands := []*exec.Cmd{build_command}

	run_command := exec.Command(config.Executable, config.Args...)
	run_command.Stdin = os.Stdin
	run_command.Stdout = os.Stdout
	run_command.Stderr = os.Stderr

	commands = append(commands, run_command)

	for _, v := range commands {
		err := v.Start()
		if err != nil {
			log.Fatalln("StartError: ", err.Error())
		}
		v.Wait()
		if err != nil {
			log.Fatalln("WaitError: ", err.Error())
		}
	}
}
