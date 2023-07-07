package tools

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"os"
	"strings"
)

var validResponse = []rune{'y', 'n'}

// GetUserConfirmation prompts user to confirm operation
// msg Message to prompt the user with
// validate returns true/false on success or terminates the process
// msg: prompt to display to the user asking for a response.
// error: error message to display if app should terminate
// terminate:  when set to true will terminate the app user response is not valid.
func GetUserConfirmation(msg, error string, terminate bool) bool {
	if error == "" {
		error = "Goodbye"
	}
	for {
		fmt.Printf(msg)
		r := bufio.NewReader(os.Stdin)
		ans, _ := r.ReadString('\n')
		ans = strings.ToLower(ans)
		if !slices.Contains(validResponse, rune(ans[0])) {
			log.Error("Invalid response, please try again.  Only [yes/no] are supported")
			continue
		}
		//Validate Response
		if ans[0] != 'y' && terminate {
			log.Fatal(error)
		} else if ans[0] != 'y' {
			return false
		}
		return true
	}

}
