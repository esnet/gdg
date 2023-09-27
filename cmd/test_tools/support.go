package test_tools

import "os"

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
