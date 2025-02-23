package cmd

import (
	"errors"
	"regexp"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	//"github.com/spf13/viper"

	"github.com/soarinferret/mcc/internal/meshrouter"
)

var routeCmd = &cobra.Command{
	Use:     "route",
	Aliases: []string{"r"},
	Short:   "Forward TCP traffic to specified Node",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {

		bindAddress, _ := cmd.Flags().GetString("bind-address")
		nodeID, _ := cmd.Flags().GetString("nodeid")
		debug, _ := cmd.Flags().GetBool("debug")

		localport, target, remoteport, err := parseBindAddress(bindAddress)
		if err != nil {
			fmt.Println("Error parsing bind address: ", err)
			return
		}

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

		meshrouter.StartRouter()

	},
}

func init() {
	rootCmd.AddCommand(routeCmd)

	routeCmd.Flags().StringP("nodeid", "i", "", "Mesh Central Node ID")
	routeCmd.Flags().StringP("bind-address", "L", "", "localport:[target:]remoteport")
	routeCmd.Flags().BoolP("debug", "", false, "Enable debug logging")
}

// parseBindAddress parses a bind address string in the format:
// "localport:target:remoteport" or "localport:remoteport"
func parseBindAddress(s string) (localPort int, target string, remotePort int, err error) {
	// Define regex pattern to match both formats
	pattern := `^(\d+)(?::([\w\.\-]+))?:(\d+)$`
	re := regexp.MustCompile(pattern)

	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return 0, "", 0, errors.New("invalid bind address format")
	}

	localPort, _ = strconv.Atoi(matches[1]) // First capture group (local port)
	target = matches[2]
	remotePort, _ = strconv.Atoi(matches[3]) // Third capture group (remote port)

	// If target is "127.0.0.1", set to nothing
	if target == "127.0.0.1" {
		target = ""
	}

	return localPort, target, remotePort, nil
}
