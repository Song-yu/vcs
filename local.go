package vcs

import (
	"time"
	"path/filepath"
	"os"
	"strings"
)

func NewLocalRepo(remote, local string) (*LocalRepo, error) {
	r := &LocalRepo{}
	remote = strings.Replace(remote, "file://", "", -1)
	r.setRemote(remote)
	r.setLocalPath(local)
	r.Logger = Logger

	return r, nil
}

type LocalRepo struct {
	base
}

func (s *LocalRepo) Vcs() Type {
	return Local
}

func (s *LocalRepo) Remote() string {
	return s.remote
}

func (s *LocalRepo) LocalPath() string {
	return s.local
}

func (s *LocalRepo) Get() error {
	out, err := s.run("cp", "-fr", s.Remote(), s.LocalPath())

	if err != nil && s.isUnableToCreateDir(err) {

		basePath := filepath.Dir(filepath.FromSlash(s.LocalPath()))
		if _, err := os.Stat(basePath); os.IsNotExist(err) {
			err = os.MkdirAll(basePath, 0755)
			if err != nil {
				return NewLocalError("Unable to create directory", err, "")
			}

			out, err = s.run("cp", "-fr", s.Remote(), s.LocalPath())
			if err != nil {
				return NewRemoteError("Unable to get repository", err, string(out))
			}
			return err
		}

	} else if err != nil {
		return NewRemoteError("Unable to get repository", err, string(out))
	}

	return nil
}

func (s *LocalRepo) Init() error {
	return nil
}

func (s *LocalRepo) Update() error {
	return nil
}

func (s *LocalRepo) UpdateVersion(string) error {
	panic("implement me")
}

func (s *LocalRepo) Version() (string, error) {
	return "Local", nil
}

func (s *LocalRepo) Current() (string, error) {
	panic("implement me")
}

func (s *LocalRepo) Date() (time.Time, error) {
	panic("implement me")
}

func (s *LocalRepo) CheckLocal() bool {
	panic("implement me")
}

func (s *LocalRepo) Branches() ([]string, error) {
	panic("implement me")
}

func (s *LocalRepo) Tags() ([]string, error) {
	panic("implement me")
}

func (s *LocalRepo) IsReference(string) bool {
	panic("implement me")
}

func (s *LocalRepo) IsDirty() bool {
	return false
}

func (s *LocalRepo) CommitInfo(string) (*CommitInfo, error) {
	ci := &CommitInfo{
		Commit:  "",
		Author:  "",
		Date:    time.Now(),
		Message: "",
	}

	return ci, nil
}

func (s *LocalRepo) TagsFromCommit(string) ([]string, error) {
	panic("implement me")
}

func (s *LocalRepo) Ping() bool {
	panic("implement me")
}

func (s *LocalRepo) ExportDir(dir string) error {
	out, err := s.RunFromDir("rsync", "-rltvp","--delete","--exclude","vendor",
		s.local+string(os.PathSeparator), dir)
	s.log(out)
	if err != nil {
		return NewLocalError("Unable to export source", err, string(out))
	}

	return nil
}

func (s *LocalRepo) isUnableToCreateDir(err error) bool {
	msg := err.Error()
	if strings.HasPrefix(msg, "could not create work tree dir") ||
		strings.HasPrefix(msg, "不能创建工作区目录") ||
		strings.HasPrefix(msg, "no s'ha pogut crear el directori d'arbre de treball") ||
		strings.HasPrefix(msg, "impossible de créer le répertoire de la copie de travail") ||
		strings.HasPrefix(msg, "kunde inte skapa arbetskatalogen") ||
		(strings.HasPrefix(msg, "Konnte Arbeitsverzeichnis") && strings.Contains(msg, "nicht erstellen")) ||
		(strings.HasPrefix(msg, "작업 디렉터리를") && strings.Contains(msg, "만들 수 없습니다")) {
		return true
	}

	return false
}
