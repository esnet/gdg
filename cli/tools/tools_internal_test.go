// tools_internal_test.go exercises small unexported helpers in tools.go that
// are not reachable from outside the package: getBasicRoles, validBasicRole,
// and needsLogin's Cobra-parent-chain walk.
package tools

import (
	"testing"

	"github.com/bep/simplecobra"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGetBasicRoles(t *testing.T) {
	roles := getBasicRoles()
	assert.Equal(t, []string{"admin", "editor", "viewer"}, roles)
}

func TestValidBasicRole(t *testing.T) {
	assert.True(t, validBasicRole("admin"))
	assert.True(t, validBasicRole("editor"))
	assert.True(t, validBasicRole("viewer"))
	assert.False(t, validBasicRole("superuser"))
	assert.False(t, validBasicRole(""))
}

// buildCommandeer wires up a minimal Cobra parent chain (tools -> group ->
// leaf) so needsLogin can walk runner.CobraCommand.Parent() the same way it
// would for a real invocation.
func buildCommandeer(groupName, leafName string) *simplecobra.Commandeer {
	root := &cobra.Command{Use: "tools"}
	group := &cobra.Command{Use: groupName}
	leaf := &cobra.Command{Use: leafName}
	root.AddCommand(group)
	group.AddCommand(leaf)
	return &simplecobra.Commandeer{CobraCommand: leaf}
}

func TestNeedsLoginFalseForNoLoginGroups(t *testing.T) {
	for _, group := range []string{"contexts", "helpers", "plugins"} {
		runner := buildCommandeer(group, "leaf")
		assert.False(t, needsLogin(runner), "group %q should not require login", group)
	}
}

func TestNeedsLoginTrueForOtherGroups(t *testing.T) {
	for _, group := range []string{"users", "auth", "organizations", "devel"} {
		runner := buildCommandeer(group, "leaf")
		assert.True(t, needsLogin(runner), "group %q should require login", group)
	}
}

func TestNeedsLoginTrueWhenNoParent(t *testing.T) {
	leaf := &cobra.Command{Use: "tools"}
	runner := &simplecobra.Commandeer{CobraCommand: leaf}
	assert.True(t, needsLogin(runner))
}
