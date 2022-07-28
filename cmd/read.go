package cmd

import (
	"MemInspector/scan"
	"encoding/hex"
	"github.com/spf13/cobra"
	"strconv"
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read <pid> <addr> [size]",
	Short: "read mem of target progress",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.ParseUint(args[0], 0, 32)
		if err != nil {
			return err
		}
		addr, err := strconv.ParseUint(args[1], 0, 32)
		if err != nil {
			return err
		}
		size := uint64(4)
		reader := scan.NewReader(uint(pid), scan.READER_IMPL_RPM)
		if len(args) == 3 {
			size, err = strconv.ParseUint(args[2], 0, 32)
			if err != nil {
				return err
			}
		}
		data, err := reader.Read(uint(addr), uint(size))
		if err != nil {
			return err
		}
		cmd.Println(hex.EncodeToString(data))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(readCmd)
	//readCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
