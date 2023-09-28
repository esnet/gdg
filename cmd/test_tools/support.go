package test_tools

import "os"

// InterceptStdout is a test helper function that will redirect all stdout in and out to a different file stream.
// It returns the stdout, stderr, and a function to be invoked to close the streams.
func InterceptStdout() (*os.File, *os.File, func()) {
	backupStd := os.Stdout
	backupErr := os.Stderr
	r, w, _ := os.Pipe()
	//Restore streams
	cleanup := func() {
		os.Stdout = backupStd
		os.Stderr = backupErr
	}
	os.Stdout = w
	os.Stderr = w

	return r, w, cleanup

}
