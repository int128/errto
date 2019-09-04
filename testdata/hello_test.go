package testdata

import "testing"

func TestHello(t *testing.T) {
	err := Hello()
	if err == nil {
		t.Errorf("err wants nil but was %+v", err)
	}
}
