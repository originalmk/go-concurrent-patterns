//go:build !validonly

package synchro

import "testing"

func TestInvalid(t *testing.T) {
	t.Run("first", func(*testing.T) { firstInvalid() })
	t.Run("second", func(*testing.T) { secondInvalid() })
	t.Run("third", func(*testing.T) { thirdInvalid() })
}
