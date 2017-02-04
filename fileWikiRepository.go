package main

import (
	"errors"
	"gopkg.in/libgit2/git2go.v25"
	"io/ioutil"
	"os"
	"strings"
)

type FileWikiRepository struct {
	Root        string
	Repo        *git.Repository
	PushOptions *git.PushOptions
}

func NewFileWikiRepository(path string, cloneFromGitRepo string, initFromGitRepo string, originGitRepo string) (*FileWikiRepository, error) {
	if initFromGitRepo != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := initialiseWikiFromGitRepository(path, initFromGitRepo)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("Path '" + path + "' already exists, unable to init from '" + initFromGitRepo + "'.")
		}
	}

	if cloneFromGitRepo != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := initialiseWikiAsCloneFromGitRepository(path, cloneFromGitRepo)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.New("Path '" + path + "' already exists, unable to clone from '" + cloneFromGitRepo + "'.")
		}
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errors.New("Path '" + path + "' does not exist, please init first.")
	}

	repo := checkForGitRepo(path)

	if originGitRepo != "" {
		//configure for push
		err := configureOriginGitRepository(repo, originGitRepo)
		if err != nil {
			return nil, err
		}
	}

	startGitWorker()

	_, pushOptions := configureOrigin(repo)
	return &FileWikiRepository{Root: path, Repo: repo, PushOptions: pushOptions}, nil
}

func pageToFilename(root string, web string, title string) string {
	return root + "/" + relativePathToPage(web, title)
}

func relativePathToPage(web string, title string) string {
	return web + "/" + title + ".md"
}

func (r *FileWikiRepository) ReadPage(web string, title string) (*Page, error) {
	filename := pageToFilename(r.Root, web, title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func (r *FileWikiRepository) WritePage(web string, p *Page) error {
	filename := pageToFilename(r.Root, web, p.Title)
	err := ioutil.WriteFile(filename, p.Body, 0644)
	if err != nil {
		return err
	}

	GitWorkQueue <- GitWork{Action: func() {commitPage(r, relativePathToPage(web, p.Title))}}

	return nil
}

func (r *FileWikiRepository) CreateWeb(web string) (*Web, error) {
	err := CopyDir(r.Root+"/_empty", r.Root+"/"+web)
	if err != nil {
		return nil, err
	}

	GitWorkQueue <- GitWork{Action: func() {commitWeb(r, web)}}

	//TODO... create web with properties?
	return &Web{Name:web}, nil
}

func (r *FileWikiRepository) LoadWebs() map[string]*Web {
	files, _ := ioutil.ReadDir(r.Root)
	m := map[string]*Web{}
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "_") || strings.HasPrefix(f.Name(), ".") {
			continue
		}
		m[f.Name()] = &Web{Name:f.Name()}
	}
	return m
}
