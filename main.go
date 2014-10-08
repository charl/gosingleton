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
func UniqueBinary(pid int) error {
	// Resolve the exe symlink for this PID.
	path, err := resolveExeSymlink(pid)
	if err != nil {
		return err
	}

	// Build an inverted index of all /proc/PID/exe symlinks.
	index, err := buildInvertedIndex()
	if err != nil {
		return err
	}

	// Query the index to see if there are multiple instances.
	bin, ok := index[path]
	if !ok {
		return fmt.Errorf("Unknown binary path %s", path)
	}

	if bin.Size() > 1 {
		return fmt.Errorf("Existing binary %s detected!", path)
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
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}

		// Skip this path if the exe symlink does not resolve properly.
		path, err := resolveExeSymlink(pid)
		if err != nil {
			continue
		}

		if bin, ok := index[path]; ok {
			bin.Add(pid)
		} else {
			index[path] = set.New(pid)
		}
	}

	return index, nil
}
