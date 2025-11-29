package util

import (
	"compress/gzip"
	"fmt"
	"os"
	"os/exec"
)

// CommandExists checks if a command is available in the PATH.
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// RunDumpToFile runs a dump command, compresses the output with gzip, and writes to a file.
// This avoids using 'sh -c' and handles piping in Go.
func RunDumpToFile(dumpCmd *exec.Cmd, filePath string) error {
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// Pipe dump command stdout to gzip writer
	dumpCmd.Stdout = gzipWriter
	
	// Capture stderr for debugging
	dumpCmd.Stderr = os.Stderr

	if err := dumpCmd.Start(); err != nil {
		return fmt.Errorf("failed to start dump command: %w", err)
	}

	if err := dumpCmd.Wait(); err != nil {
		return fmt.Errorf("dump command failed: %w", err)
	}
    
    // Ensure gzip is closed before file close (handled by defer, but we need to check error if flush fails)
    if err := gzipWriter.Close(); err != nil {
        return fmt.Errorf("failed to close gzip writer: %w", err)
    }

	return nil
}

// RestoreFromFile runs a restore command, reading from a gzipped file.
func RestoreFromFile(restoreCmd *exec.Cmd, filePath string) error {
    inFile, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("failed to open backup file: %w", err)
    }
    defer inFile.Close()
    
    gzipReader, err := gzip.NewReader(inFile)
    if err != nil {
        return fmt.Errorf("failed to create gzip reader: %w", err)
    }
    defer gzipReader.Close()
    
    restoreCmd.Stdin = gzipReader
    restoreCmd.Stderr = os.Stderr
    
    if err := restoreCmd.Start(); err != nil {
        return fmt.Errorf("failed to start restore command: %w", err)
    }
    
    if err := restoreCmd.Wait(); err != nil {
        return fmt.Errorf("restore command failed: %w", err)
    }
    
    return nil
}
