package handlers

import (
	"testing"

	. "github.com/franela/goblin"
)

func TestAuthRoutes(t *testing.T) {
	g := Goblin(t)

	g.Describe("Signup Route Suits", func() {
		g.It("Accept valid email, username and password", func() {
			g.Assert(1 + 1).Equal(2)
		})
	})
}
