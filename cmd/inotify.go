package cmd

import (
	"MemInspector/scan"
	"fmt"
	"github.com/spf13/cobra"
)

// inotifyCmd represents the inotify command
var inotifyCmd = &cobra.Command{
	Use:   "inotify <on|off|check>",
	Short: "enable or disable inotify",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "on":
			err := scan.EnableInotify()
			return err
		case "off":
			err := scan.DisableInotify()
			return err
		case "check":
			stat, err := scan.CheckInotify()
			cmd.Printf("Inotify status: %v\n", stat)
			return err
		default:
			return fmt.Errorf("invalid argument: %s", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(inotifyCmd)
}
