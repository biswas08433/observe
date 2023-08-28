package main

import (
	"container/list"
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

		if len(config.Executable) == 0 {
			log.Fatalln("No run command")
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
		go EnQueue(watcher, 300, pipe)
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
			if timeout && lastEvent.Has(fsnotify.Write) {
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

func ConnectIO(cmd *exec.Cmd, in, out, err *os.File) {
	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = err
}

// This function fires when there is an event in the local queue
func ProcessQueue(pipe chan fsnotify.Event, config *ServerConfig) {
	var event fsnotify.Event
	command_buffer := list.New()
	command_buffer.Init()
	PushCommand(config, command_buffer)
	PushCommand(config, command_buffer)
	ExecuteCommand(command_buffer)
	for {
		event = <-pipe
		log.Println(event)

		iter := command_buffer.Front()
		command_buffer.Remove(iter)
		iter = command_buffer.Front()

		cmd := iter.Value.(*exec.Cmd)
		err := cmd.Process.Kill()
		if err != nil {
			log.Fatalln("Kill Error: ", err.Error(), cmd)
		}
		// log.Println("Killed: ", cmd, cmd.Process)
		err = cmd.Wait()
		if err != nil {
			log.Println("Wait error: ", err.Error())
		}
		// log.Println("Released: ", cmd, cmd.Process)
		command_buffer.Remove(iter)
		// log.Println("Removed: ", )
		log.Println("Restarting...")
		ExecuteCommand(command_buffer)
		// log.Println("Executed")
		PushCommand(config, command_buffer)
	}
}

func PushCommand(config *ServerConfig, command_buffer *list.List) {

	build_command := exec.Command(config.BuildSystem, config.BuildArgs...)
	ConnectIO(build_command, os.Stdin, os.Stdout, os.Stderr)
	run_command := exec.Command(config.Executable, config.Args...)
	ConnectIO(run_command, os.Stdin, os.Stdout, os.Stderr)

	command_buffer.PushBack(build_command)
	command_buffer.PushBack(run_command)
}

func ExecuteCommand(command_buffer *list.List) {
	var iter *list.Element
	iter = command_buffer.Front()
	for i := 0; i < 2; i++ {
		cmd := iter.Value.(*exec.Cmd)
		err := cmd.Start()
		if err != nil {
			log.Fatalln(cmd, "StartError: ", err.Error())
		}
		// log.Println("Execution Started: ", cmd, cmd.Process)
		if i == 0 {
			cmd.Wait()
		}
		iter = iter.Next()
	}
}
