package generator

import (
	"net/url"
	"path/filepath"
	"regexp"
)

func parseHttpUrl(src string) *url.URL {
	u, err := url.Parse(src)
	if err != nil {
		return nil
	}
	return u
}

// sshPattern matches SCP-like SSH patterns (user@host:path)
var sshPattern = regexp.MustCompile("^(?:([^@]+)@)?([^:]+):/?(.+)$")

func parseSshUrl(src string) *url.URL {
	matched := sshPattern.FindStringSubmatch(src)
	if matched == nil {
		return nil
	}

	user := matched[1]
	host := matched[2]
	path := matched[3]

	return &url.URL{
		Scheme: "ssh",
		User:   url.User(user),
		Host:   host,
		Path:   path,
	}
}

func urlBaseName(u *url.URL) string {
	base := filepath.Base(u.Path)
	extension := filepath.Ext(base)
	return base[0 : len(base)-len(extension)]
}
