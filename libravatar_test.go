// Copyright 2016 by Sandro Santilli <strk@kbt.io>
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package libravatar

import "testing"

func TestFromEmail(t *testing.T) {

	cases := []struct{ in, want string }{
		{"strk@kbt.io", "http://_avatars._tcp.kbt.io.avatars.kbt.io/avatar/fe2a9e759730ee64c44bf8901bf4ccc3"},
		{"strk@keybit.net", "http://cdn.libravatar.org/avatar/34bafd290f6f39380f5f87e0122daf83"},
		{"strk@nonexistent.domain", "http://cdn.libravatar.org/avatar/3f30177111597990b15f8421eaf736c7"},
	}

	cases_invalid := []struct {
		in string
	}{
		{"invalid"},
		{"invalid@"},
		{"@invalid"},
		{""},
	}

	l := New()

	for _, c := range cases_invalid {
		got, err := l.FromEmail(c.in)
		if err == nil {
			t.Errorf("fromEmail(%q) did not raise an expected error but returned %q", c.in, got)
		}
	}

	for _, c := range cases {
		got, err := l.FromEmail(c.in)
		if err != nil {
			t.Errorf("fromEmail(%q) raised an error: %q", c.in, err)
		} else if got != c.want {
			t.Errorf("fromEmail(%q) == %q, expected %q", c.in, got, c.want)
		}
	}

	// TODO: test https and parameters

}
