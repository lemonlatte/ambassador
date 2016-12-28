package ambassador

import (
	"testing"
)

func TestAmbassadorNew(t *testing.T) {
	a := New("facebook", "test-token", nil)
	t.Log(a)
}
