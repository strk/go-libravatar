// Copyright 2016 by Sandro Santilli <strk@kbt.io>
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package libravatar

import "testing"

func TestFromEmail(t *testing.T) {

	avt := New()

	// Email tests

	cases := []struct{ in, want string }{
		{"strk@kbt.io", "http://avatars.kbt.io/avatar/fe2a9e759730ee64c44bf8901bf4ccc3"},
		{"strk@keybit.net", "http://cdn.libravatar.org/avatar/34bafd290f6f39380f5f87e0122daf83"},
		{"strk@nonexistent.domain", "http://cdn.libravatar.org/avatar/3f30177111597990b15f8421eaf736c7"},
		{"invalid", ""},
		{"invalid@", ""},
		{"@invalid", ""},
	}

	for _, c := range cases {
		got, err := avt.FromEmail(c.in)
		if err != nil {
			if c.want == "" {
				t.Skipf("Forced error successfully (%s)", err)
				continue
			}
			got = err.Error()
		}
		if got != c.want {
			t.Errorf("fromEmail(%q) == %q, expected %q", c.in, got, c.want)
		}
	}

	// TODO: test https with email

	// OpenID tests

	cases = []struct{ in, want string }{
		{"https://strk.kbt.io/openid/", "https://cdn.libravatar.org/avatar/1eaf3174c95d0df02f177f7f6a1df5125ad3d6603fbd062defecd30810a0463c"},
		{"invalid", ""},
	}

	for _, c := range cases {
		got, err := avt.FromURL(c.in)
		if err != nil {
			if c.want == "" {
				t.Skipf("Forced error successfully (%s)", err)
				continue
			}
			got = err.Error()
		}
		if got != c.want {
			t.Errorf("fromURL(%q) == %q, expected %q", c.in, got, c.want)
		}
	}

	// TODO: test parameters

}
