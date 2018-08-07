package gosingleton

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

// const version = "v0.0.5"

// Check if we're the only process running with this path.
func UniquePath(path string) bool {
	return false // not implemented.
}

// Check if we're the only process running with this name.
func UniqueName(path string) bool {
	return false // not implemented
}

// Check if an instance of the same binary is running.
func UniqueBinary(name string, pid int) error {
	// This was only developed for Linux so exit early if we're running in some other OS.
	if runtime.GOOS != "linux" {
		return resolveUsingPs(name)
	}

	// Resolve the exe symlink for this PID.
	path, err := resolveExeSymlink(pid)
	if err != nil {
		return err
	}

	// Query the /proc fs to see if there are multiple instances.
	err = existingProcEntry(path)
	if err == nil {
		return fmt.Errorf("Unknown binary path %s", path)
	}
	if err != nil {
		return fmt.Errorf("Existing binary %s detected!", path)
	}

	return nil
}

// resolveUsingPs does a UNIX _ps -axw | grep $name | grep -v grep_ to check if the named proc is running.
func resolveUsingPs(name string) error {
	var b bytes.Buffer

	c1 := exec.Command("ps", "-axw")
	c2 := exec.Command("grep", name)
	c3 := exec.Command("grep", "-v", "grep")

	c3.Stdin, _ = c2.StdoutPipe()
	c2.Stdin, _ = c1.StdoutPipe()
	c3.Stdout = os.Stdout

	_ = c3.Start()
	_ = c2.Start()
	_ = c1.Run()
	_ = c2.Wait()
	_ = c2.Run()
	_ = c3.Wait()

	_, err := io.Copy(os.Stdout, &b)
	if err != nil {
		return err
	}

	if b.String() != "" {
		return fmt.Errorf("Existing binary %s detected!", name)
	}

	return nil
}

// Resole the exe symlink for a PID.
func resolveExeSymlink(pid int) (string, error) {
	path, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		return "", err
	}

	return path, nil
}

// existingProcEntry checks to see if this path is already represented in the /proc fs.
func existingProcEntry(path string) error {
	index := make(map[string]int)

	// Find all entries under /proc.
	entries, err := ioutil.ReadDir("/proc")
	if err != nil {
		return err
	}

	// Filter for PIDs and resolve their exe symlinks.
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}

		// Skip this path if the exe symlink does not resolve properly.
		resolvedPath, err := resolveExeSymlink(pid)
		if err != nil {
			continue
		}

		index[resolvedPath]++
	}

	c := index[path]
	if c > 1 {
		return errors.New("multiple instnaces")
	}

	return nil
}
