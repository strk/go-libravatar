// Copyright 2016 by Sandro Santilli <strk@kbt.io>
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Implements support for federated avatars lookup.
// See https://wiki.libravatar.org/api/
package libravatar

import (
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
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

/* This should be moved in its own file */
type cacheKey struct {
	service string
	domain  string
}

type cacheValue struct {
	target    string
	checkedAt time.Time
}

type Libravatar struct {
	defUrl            string // default url
	picSize           int    // picture size
	useDomain         bool   //
	useHTTPS          bool
	fallbackHost      string
	nameCache         map[cacheKey]cacheValue
	nameCacheDuration time.Duration
}

// Instanciate a library handle
func New() *Libravatar {
	o := &Libravatar{fallbackHost: "cdn.libravatar.org"}
	o.nameCache = make(map[cacheKey]cacheValue)
	// According to https://wiki.libravatar.org/running_your_own/
	// the time-to-live (cache expiry) should be set to at least 1 day.
	o.nameCacheDuration = 24 * time.Hour
	return o
}

func getSHA256(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	hash := sha256.New()
	hash.Write([]byte(s))
	return fmt.Sprintf("%x", hash.Sum(nil))
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

// Set useHTTPS flag (only used with email)
func (v *Libravatar) SetUseHTTPS(use bool) {
	v.useHTTPS = use
}

func (v *Libravatar) baseURL(domain string, useHTTPS bool) (url string, err error) {
	//(cname string, addrs []*SRV, err error)
	//var cname, addrs, err string
	service := ""
	if useHTTPS {
		url = "https://"
		service = "avatars-sec"
	} else {
		url = "http://"
		service = "avatars"
	}
	key := cacheKey{service, domain}
	//fmt.Println("DEBUG: key is ", key)
	var tgt string
	now := time.Now()
	val, found := v.nameCache[key]
	if found && now.Sub(val.checkedAt) <= v.nameCacheDuration {
		tgt = val.target
		//fmt.Println("DEBUG: cache hit, target for ", key, " is ", tgt)
	} else {
		_, addrs, e := net.LookupSRV(service, "tcp", domain)
		fmt.Println("DEBUG: lookup of ", service, domain, " took ", time.Since(now))
		if e != nil {
			//fmt.Println("DEBUG: Lookup error: ", e) // debug
			e := e.(*net.DNSError)
			if e.IsTimeout {
				// A timeout we'll consider a real error
				// TODO: use fallback instead i
				err = e
				fmt.Println("DEBUG: lookup timeout") // debug
				return
			}
			// An host-not-found (assumed here) would trigger a fallback
			// to libravatar.org
			url += v.fallbackHost
			v.nameCache[key] = cacheValue{checkedAt: now, target: v.fallbackHost}
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
		tgt = addrs[0].Target
		tgt = tgt[:len(tgt)-1] // strip the ending dot
		v.nameCache[key] = cacheValue{checkedAt: now, target: tgt}
	}
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
	baseurl, err := v.baseURL(domain, v.useHTTPS)
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

// Return url of the avatar for the given url (tipically for OpenID)
func (v *Libravatar) FromURL(in string) (url string, err error) {
  parts := strings.Split(in, "/")
	// http://domain <-- smallest valid, has 3 parts
  if len(parts) < 3 {
		err = errors.New("invalid url")
		return
	}

	domain := parts[2]
	if len(domain) == 0 {
		err = errors.New("invalid email")
		return
	}

	var useHTTPS = v.useHTTPS
  if parts[0] == "https:" {
      useHTTPS = true;
  }

	hash := getSHA256(in)
	baseurl, err := v.baseURL(domain, useHTTPS)
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
