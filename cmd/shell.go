
package cmd

import (
	"github.com/spf13/cobra"
	//"github.com/spf13/viper"

	"github.com/soarinferret/mcc/internal/meshcentral"
)


var shellCmd = &cobra.Command{
	Use:     "shell",
	Short:   "Opens a root shell directly to the node",
	Long:    ``,
	Run: func(cmd *cobra.Command, args []string) {

		nodeID, _ := cmd.Flags().GetString("nodeid")
		debug, _ := cmd.Flags().GetBool("debug")
		powershell, _ := cmd.Flags().GetBool("powershell")


		meshcentral.ApplySettings(
			nodeID,
			0,
			0,
			"",
			debug,
		)

		meshcentral.StartSocket()

		if nodeID == "" {
			devices := meshcentral.GetDevices()
			filterAndSortDevices(&devices)
			nodeID = searchDevices(&devices)

			meshcentral.ApplySettings(
				nodeID,
				0,
				0,
				"",
				debug,
			)
		}

		//ready := make(chan struct{})

		// open shell
		protocol := 1
		if powershell {
			protocol = 6
		}
		meshcentral.StartShell(protocol)

		meshcentral.StopSocket()

	},
}
func init() {
	rootCmd.AddCommand(shellCmd)

	shellCmd.Flags().StringP("nodeid", "i", "", "Mesh Central Node ID")
	shellCmd.Flags().BoolP("debug", "", false, "Enable debug logging")
	shellCmd.Flags().BoolP("powershell", "p", false, "Use powershell instead of cmd.exe (windows agents only")
}
