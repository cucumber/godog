package step_func

import (
	"fmt"
)

// Godogs is an example behavior holder.
type Godogs int

// Add increments Godogs count.
func (g *Godogs) Add(n int) {
	*g = *g + Godogs(n)
}

// Eat decrements Godogs count or fails if there is not enough available.
func (g *Godogs) Eat(n int) error {
	ng := Godogs(n)

	if (g == nil && ng > 0) || ng > *g {
		return fmt.Errorf("you cannot eat %d godogs, there are %d available", n, g.Available())
	}

	if ng > 0 {
		*g = *g - ng
	}

	return nil
}

// Available returns the number of currently available Godogs.
func (g *Godogs) Available() int {
	if g == nil {
		return 0
	}

	return int(*g)
}
