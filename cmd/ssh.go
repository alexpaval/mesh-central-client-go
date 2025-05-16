
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"github.com/spf13/cobra"
	//"github.com/spf13/viper"

	"github.com/soarinferret/mcc/internal/meshrouter"
)


var sshCmd = &cobra.Command{
	Use:     "ssh [user][@target]",
	Short:   "Shortcut to ssh into a node",
	Long:    `Opens SSH connection with the OpenSSH Client to a node via the local proxy`,
	Run: func(cmd *cobra.Command, args []string) {

		user := "root"
		target := ""

		if len(args) == 1 {
			// parse user@target
			parts := strings.Split(args[0], "@")
			user = parts[0]
			if len(parts) == 2 {
				target = parts[1]
			}
		}

		remoteport, _ := cmd.Flags().GetInt("port")

		nodeID, _ := cmd.Flags().GetString("nodeid")
		debug, _ := cmd.Flags().GetBool("debug")

		// generate random local port num
		localport := 0

		meshrouter.ApplySettings(
			nodeID,
			remoteport,
			localport,
			target,
			debug,
		)

		meshrouter.StartSocket()

		if nodeID == "" {
			devices := meshrouter.GetDevices()
			filterAndSortDevices(&devices)
			nodeID = searchDevices(&devices)

			meshrouter.ApplySettings(
				nodeID,
				remoteport,
				localport,
				target,
				debug,
			)
		}

		ready := make(chan struct{})

		// start proxy
		go meshrouter.StartRouter(ready)

		// wait for proxy to be ready
		<-ready

		// start ssh client
		sshPort := meshrouter.GetLocalPort()
		fmt.Printf("SSH into %s:%d via 127.0.0.1%d\n", target, remoteport, sshPort)
		sshCmd := exec.Command(	"ssh", "-o", "ServerAliveInterval=60",
								"-o", "ServerAliveCountMax=3",
							 	"-o", "StrictHostKeyChecking=no",
								"-o", "UserKnownHostsFile=/dev/null",
							 	fmt.Sprintf("-p%d", sshPort), fmt.Sprintf("%s@127.0.0.1", user),
		)
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr
		sshCmd.Stdin = os.Stdin
		err := sshCmd.Run()
		if err != nil {
			fmt.Printf("Unable to start SSH client: %v\n", err)
		}


	},
}
func init() {
	rootCmd.AddCommand(sshCmd)

	sshCmd.Flags().StringP("nodeid", "i", "", "Mesh Central Node ID")
	sshCmd.Flags().IntP("port", "p", 22, "Define the remote ssh port")
	sshCmd.Flags().BoolP("debug", "", false, "Enable debug logging")
}
