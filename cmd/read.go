package cmd

import (
	"MemInspector/scan"
	"encoding/hex"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

var (
	hexdump  *bool
	str      *bool
	anti     *bool
	format   *string
	filename *string
)

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read <pid> <addr> [size]",
	Short: "read mem of target progress",
	Args:  cobra.RangeArgs(2, 3),
	RunE:  doCommand,
}

func doCommand(cmd *cobra.Command, args []string) error {
	pid, err := strconv.ParseUint(args[0], 0, 32)
	if err != nil {
		return err
	}
	addr, err := strconv.ParseUint(args[1], 0, 64)
	if err != nil {
		return err
	}
	size := uint64(4)
	var reader *scan.Reader
	if *anti {
		reader = scan.NewReader(uint(pid), scan.READER_IMPL_RPM_WITH_ANTI)
	} else {
		reader = scan.NewReader(uint(pid), scan.READER_IMPL_RPM)
	}
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
	if *filename != "" {
		err = os.WriteFile(*filename, data, 0644)
		if err != nil {
			return err
		}
	}
	if *hexdump {
		cmd.Println(hex.Dump(data))
		return nil
	}
	if *str {
		cmd.Println(string(data)) //encoding may be wrong
		return nil
	}
	if *format != "" {
		cmd.Printf(*format+"\n", data)
		return nil
	}
	cmd.Println(hex.EncodeToString(data))
	return nil
}

func init() {
	rootCmd.AddCommand(readCmd)
	hexdump = readCmd.Flags().Bool("hex", false, "output in hexdump style")
	str = readCmd.Flags().Bool("str", false, "convert to string")
	anti = readCmd.Flags().BoolP("anti", "a", false, "anti page status detection")
	format = readCmd.Flags().StringP("format", "f", "", "print by custom format")
	filename = readCmd.Flags().StringP("file", "o", "", "output to file")
	readCmd.MarkFlagsMutuallyExclusive("hex", "str", "format")
	//readCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
