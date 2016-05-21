// Copyright 2016 by Sandro Santilli <strk@kbt.io>
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Implements support for federated avatars lookup.
// See https://wiki.libravatar.org/api/
package libravatar

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net"
	//"net/http"
	//"net/url"
	"strings"
)

// Default images (to be used as defaultURL)
const (
	// Do not load any image if none is associated with the email
	// hash, instead return an HTTP 404 (File Not Found) response
	HTTP404 = "404"
	// (mystery-man) a simple, cartoon-style silhouetted outline of
	// a person (does not vary by email hash)
	MysteryMan = "mm"
	// a geometric pattern based on an email hash
	IdentIcon = "identicon"
	// a generated 'monster' with different colors, faces, etc
	MonsterID = "monsterid"
	// generated faces with differing features and backgrounds
	Wavatar = "wavatar"
	// awesome generated, 8-bit arcade-style pixelated faces
	Retro = "retro"
)

type Libravatar struct {
	defUrl       string // default url
	picSize      int    // picture size
	useDomain    bool   //
	useHTTPS     bool
	fallbackHost string
}

// Instanciate a library handle
func New() *Libravatar {
	o := &Libravatar{fallbackHost: "cdn.libravatar.org"}
	return o
}

func getMD5(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	hash := md5.New()
	hash.Write([]byte(s))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// Set the hostname for fallbacks in case no avatar service is defined
// for a domain
func (v *Libravatar) SetFallbackHost(host string) {
	v.fallbackHost = host
}

// Set useHTTPS flag
func (v *Libravatar) SetUseHTTPS(use bool) {
	v.useHTTPS = use
}

func (v *Libravatar) baseURL(domain string) (url string, err error) {
	//(cname string, addrs []*SRV, err error)
	//var cname, addrs, err string
	service := ""
	if v.useHTTPS {
		url = "https://"
		service = "avatars-sec"
	} else {
		url = "http://"
		service = "avatars"
	}
	_, addrs, e := net.LookupSRV(service, "tcp", domain)
	if e != nil {
		//fmt.Printf("DEBUG: Lookup error: %s\n", err) // debug
		e := e.(*net.DNSError)
		if e.IsTimeout {
			// A timeout we'll consider a real error
			err = e
			return
		}
		// An host-not-found (assumed here) would trigger a fallback
		// to libravatar.org
		url += v.fallbackHost
		return
	}
	// TODO: sort by priority, but for now just pick the first
	if len(addrs) < 1 {
		err = fmt.Errorf("empty SRV response")
		return
	}
	/*
		for _, v := range addrs {
			fmt.Printf("DEBUG: Target:%s - Port:%d - Priority:%d - Weight:%d\n",
				v.Target, v.Port, v.Priority, v.Weight)
		}
	*/
	tgt := addrs[0].Target
	tgt = tgt[:len(tgt)-1] // strip the ending dot
	url += tgt
	return
}

// Return url of the avatar for the given email
func (v *Libravatar) FromEmail(email string) (url string, err error) {
	i := strings.Index(email, "@")
	if i < 1 {
		err = errors.New("invalid email")
		return
	}
	domain := email[i+1:]
	if len(domain) == 0 {
		err = errors.New("invalid email")
		return
	}
	hash := getMD5(email)
	baseurl, err := v.baseURL(domain)
	if err != nil {
		return
	}
	url = baseurl + "/avatar/" + hash

	// Append parameters as needed
	sep := "?"
	if v.useDomain {
		url += sep + "domain=" + domain
		sep = "&"
	}
	if v.picSize != 0 {
		url += sep + "s=" + string(v.picSize)
		sep = "&"
	}
	if v.defUrl != "" {
		url += sep + "d=" + v.defUrl
		sep = "&"
	}
	return
}
