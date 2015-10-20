package vcs

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var hgDetectURL = regexp.MustCompile("default = (?P<foo>.+)\n")

// NewHgRepo creates a new instance of HgRepo. The remote and local directories
// need to be passed in.
func NewHgRepo(remote, local string) (*HgRepo, error) {
	ltype, err := DetectVcsFromFS(local)

	// Found a VCS other than Hg. Need to report an error.
	if err == nil && ltype != Hg {
		return nil, ErrWrongVCS
	}

	r := &HgRepo{}
	r.setRemote(remote)
	r.setLocalPath(local)
	r.Logger = Logger

	// Make sure the local Hg repo is configured the same as the remote when
	// A remote value was passed in.
	if err == nil && r.CheckLocal() == true {
		// An Hg repo was found so test that the URL there matches
		// the repo passed in here.
		c := exec.Command("hg", "paths")
		c.Dir = local
		c.Env = envForDir(c.Dir)
		out, err := c.CombinedOutput()
		if err != nil {
			return nil, err
		}

		m := hgDetectURL.FindStringSubmatch(string(out))
		if m[1] != "" && m[1] != remote {
			return nil, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Hg repo use that one.
		if remote == "" && m[1] != "" {
			r.setRemote(m[1])
		}
	}

	return r, nil
}

// HgRepo implements the Repo interface for the Mercurial source control.
type HgRepo struct {
	base
}

// Vcs retrieves the underlying VCS being implemented.
func (s HgRepo) Vcs() Type {
	return Hg
}

// Get is used to perform an initial clone of a repository.
func (s *HgRepo) Get() error {
	_, err := s.run("hg", "clone", s.Remote(), s.LocalPath())
	return err
}

// Update performs a Mercurial pull to an existing checkout.
func (s *HgRepo) Update() error {
	_, err := s.runFromDir("hg", "update")
	return err
}

// UpdateVersion sets the version of a package currently checked out via Hg.
func (s *HgRepo) UpdateVersion(version string) error {
	_, err := s.runFromDir("hg", "pull")
	if err != nil {
		return err
	}
	_, err = s.runFromDir("hg", "update", version)
	return err
}

// Version retrieves the current version.
func (s *HgRepo) Version() (string, error) {
	out, err := s.runFromDir("hg", "identify")
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(string(out), " ", 2)
	sha := parts[0]
	return strings.TrimSpace(sha), nil
}

// CheckLocal verifies the local location is a Git repo.
func (s *HgRepo) CheckLocal() bool {
	if _, err := os.Stat(s.LocalPath() + "/.hg"); err == nil {
		return true
	}

	return false
}

// Branches returns a list of available branches
func (s *HgRepo) Branches() ([]string, error) {
	out, err := s.runFromDir("hg", "branches")
	if err != nil {
		return []string{}, err
	}
	branches := s.referenceList(string(out), `(?m-s)^(\S+)`)
	return branches, nil
}

// Tags returns a list of available tags
func (s *HgRepo) Tags() ([]string, error) {
	out, err := s.runFromDir("hg", "tags")
	if err != nil {
		return []string{}, err
	}
	tags := s.referenceList(string(out), `(?m-s)^(\S+)`)
	return tags, nil
}

// IsReference returns if a string is a reference. A reference can be a
// commit id, branch, or tag.
func (s *HgRepo) IsReference(r string) bool {
	_, err := s.runFromDir("hg", "log", "-r", r)
	if err == nil {
		return true
	}

	return false
}
