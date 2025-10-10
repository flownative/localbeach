// Copyright 2019-2025 Robert Lemke, Karsten Dambekalns, Christian Müller
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/flownative/localbeach/pkg/beachsandbox"
	"github.com/flownative/localbeach/pkg/exec"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:                "exec",
	Short:              "Execute a command in or enter a Local Beach container",
	Long:               "",
	DisableFlagParsing: true,
	Run:                handleExecRun,
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func handleExecRun(cmd *cobra.Command, args []string) {
	sandbox, err := beachsandbox.GetActiveSandbox()
	if err != nil {
		log.Fatal("Could not activate sandbox: ", err)
		return
	}

	// Check if stdin is a TTY using syscall (more reliable than Mode check)
	var termios syscall.Termios
	_, _, errno := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCGETA, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	isTTY := errno == 0

	// Build Docker exec command with appropriate flags
	commandArgs := []string{"exec"}
	if isTTY {
		commandArgs = append(commandArgs, "-t", "-i")
	}
	// Note: No -i flag when not TTY since stdin isn't connected in RunCommand
	commandArgs = append(commandArgs, sandbox.ProjectName+"_php")
	if len(args) > 0 {
		commandArgs = append(commandArgs, "bash", "-l", "-c", strings.Trim(fmt.Sprint(args), "[]"))
	} else {
		commandArgs = append(commandArgs, "bash")
	}

	// Use the appropriate execution method based on TTY detection
	if isTTY {
		err = exec.RunInteractiveCommand("docker", commandArgs)
		if err != nil {
			log.Fatal(err)
			return
		}
	} else {
		output, err := exec.RunCommand("docker", commandArgs)
		fmt.Print(output)
		if err != nil {
			os.Exit(1)
		}
	}
}
