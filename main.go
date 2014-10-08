package gosingleton

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"gopkg.in/fatih/set.v0"
)

// Check if we're the only process running with this path.
func UniquePath(path string) bool {
	return false // not implemented.
}

// Check if we're the only process running with this name.
func UniqueName(path string) bool {
	return false // not implemented
}

// Check if an instance of the same binary is running.
func UniqueBinary(pid int64) (bool, error) {
	// Resolve the exe symlink for this PID.
	path, err := resolveExeSymlink(pid)
	if err != nil {
		return false, err
	}

	// Build an inverted index of all /proc/PID/exe symlinks.
	index, err := buildInvertedIndex()
	if err != nil {
		return false, err
	}

	// Query the index to see if there are multiple instances.
	bin, ok := index[path]
	if !ok {
		return false, fmt.Errorf("Unknown binary path %s", path)
	}

	return bin.Has(pid), nil
}

// Resole the exe symlink for a PID.
func resolveExeSymlink(pid int64) (string, error) {
	path, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		return "", err
	}

	return path, nil
}

// Build an inverted index of paths as keys with sets of PIDs as values.
func buildInvertedIndex() (map[string]*set.Set, error) {
	index := make(map[string]*set.Set)

	// Find all entries under /proc.
	entries, err := ioutil.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	// Filter for PIDs and resolve their exe symlinks.
	for _, e := range entries {
		pid, err := strconv.ParseInt(e.Name(), 10, 64)
		if err != nil {
			continue
		}

		path, err := resolveExeSymlink(pid)
		if err != nil {
			return nil, err
		}

		if bin, ok := index[path]; ok {
			bin.Add(pid)
		} else {
			index[path] = set.New(pid)
		}
	}

	return index, nil
}
