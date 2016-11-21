package aurora

import "testing"

func TestOutputPayloadString(t *testing.T) {
	p := &outputPayload{}
	if str := p.String(); str != "00 00 00 00 00 00 00 00 (0)" {
		t.Errorf("Unexpected string: %s", str)
	}
}

func TestInputPayloadString(t *testing.T) {
	p := &inputPayload{}
	if str := p.String(); str != "00 00 00 00 00 00 (0)" {
		t.Errorf("Unexpected string: %s", str)
	}
}
