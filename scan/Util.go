package scan

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func FindLibBase(pid, libName string) (int64, error) {
	command := "cat /proc/" + pid + "/maps|grep " + libName + "|head -n 1"
	//fmt.Println(command)
	cmd := exec.Command("su", "-c", command)
	bytes, err := cmd.Output()
	if err != nil {
		fmt.Printf("cannot read maps file")
		return -1, err
	}
	if string(bytes) == "" {
		return -1, errors.New("cannot find lib")
	}
	if s, err := strconv.ParseInt(strings.Split(strings.Split(string(bytes), " ")[0], "-")[0], 16, 64); err == nil {
		return s, nil
	}
	errStr := fmt.Sprintf("cannot read addr")
	err = errors.New(errStr)
	return -1, err
}

func FindLibInfo(pid, libName string) (string, error) {
	command := "cat /proc/" + pid + "/maps|grep " + libName
	//fmt.Println(command)
	cmd := exec.Command("su", "-c", command)
	bytes, err := cmd.Output()
	if err != nil {
		fmt.Printf("cannot read maps data, wrong pid or lib name\n")
		return "", err
	}
	return string(bytes), nil
}

// DisableInotify TODO: find a more elegant way to restrict app's inotify
func DisableInotify() {
	command := "echo 0 > /proc/sys/fs/inotify/max_user_watches"
	exec.Command("su", "-c", command)
}
