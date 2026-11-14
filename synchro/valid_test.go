package synchro

import "testing"

func TestValidMutex(t *testing.T) {
	t.Run("first", func(*testing.T) { firstValidMutex() })
	t.Run("second", func(*testing.T) { secondValidMutex() })
	t.Run("third", func(*testing.T) { thirdValidMutex() })
}

func TestValidChannel(t *testing.T) {
	t.Run("first", func(*testing.T) { firstValidChannel() })
	t.Run("second", func(*testing.T) { secondValidChannel() })
	t.Run("third", func(*testing.T) { thirdValidChannel() })
}

func TestValidAtomics(t *testing.T) {
	t.Run("first", func(*testing.T) { firstValidAtomics() })
}
