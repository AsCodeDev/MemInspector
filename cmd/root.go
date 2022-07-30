package cmd

import (
	"MemInspector/scan"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "MemInspector",
	Short: "A Memory Inspector which can read other processes' memory",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	if os.Getuid() != 0 {
		log.Fatalln("Running without root is not supported temporarily")
	}
	/*
		As a temporary solution, inotify will be disabled at startup by setting 'max_user_watches' to 0.
		Notice: You should start app after setting,because no progress can register new inotify watcher.
		'max_user_watches' will not restore automatically.
		You can manually set 'max_user_watches' to a normal value if you want to use inotify after using MemInspector.
		Restarting the device is also effective.
	*/
	scan.DisableInotify()
}
