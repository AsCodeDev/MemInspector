package cmd

import (
	"MemInspector/scan"
	"github.com/spf13/cobra"
)

var (
	detail *bool
)

// findlibCmd represents the findlib command
var findlibCmd = &cobra.Command{
	Use:   "findlib <pid> <libName>",
	Short: "find address of library in target progress memory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if *detail {
			info, err := scan.FindLibInfo(args[0], args[1])
			if err != nil {
				return err
			}
			cmd.Println(info)
		} else {
			addr, err := scan.FindLibBase(args[0], args[1])
			if err != nil {
				return err
			}
			cmd.Printf("Base Address of %s: %x", args[0], addr)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(findlibCmd)
	detail = findlibCmd.Flags().BoolP("detail", "d", false, "show detail")
}
