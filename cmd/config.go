package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/soarinferret/mcc/internal/config"
)

var configCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"c"},
	Short:   "Return Config Path",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println(config.GetConfigPath())

	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
