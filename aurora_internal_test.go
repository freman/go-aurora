// Copyright 2016 Shannon Wynter. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package aurora

import "testing"

func TestCalculateCRC(t *testing.T) {
	tests := []struct {
		Data   []byte
		Expect uint16
	}{
		{Data: []byte{32, 32, 32, 32, 32, 32, 32, 32}, Expect: 15784},
		{Data: []byte{2, 56, 32, 32, 32, 32, 32, 32}, Expect: 60178},
		{Data: []byte{2, 56, 1, 2, 3, 4, 5, 6}, Expect: 53051},
	}

	for _, test := range tests {
		crc := calculateCRC(test.Data)
		if crc != test.Expect {
			t.Errorf("calculateCRC(%v) = %d, expected %d", test.Data, crc, test.Expect)
		}
	}
}
