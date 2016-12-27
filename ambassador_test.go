package ambassador

import (
	"testing"
)

func TestAmbassadorNew(t *testing.T) {
	a := New("facebook", "test-token")
	t.Log(a)
}
