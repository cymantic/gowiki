package main

import (
	"os"
	"io/ioutil"
	"gopkg.in/libgit2/git2go.v25"
	log "github.com/Sirupsen/logrus"
	"time"
	"errors"
)

type FilePageStorage struct {
	Root        string
	Repo        *git.Repository
	Origin      *git.Remote
	PushOptions *git.PushOptions
}

func NewFilePageStorage(path string, cloneFromGitRepo string, initFromGitRepo string, originGitRepo string) (*FilePageStorage, error) {
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

	origin, pushOptions := configureOrigin(repo)
	return &FilePageStorage{Root:path, Repo:repo, Origin:origin, PushOptions:pushOptions}, nil
}

func pageToFilename(root string, web string, title string) string {
	return root + "/" + relativePathToPage(web, title)
}

func relativePathToPage(web string, title string) string {
	return web + "/" + title + ".md"
}

func (s *FilePageStorage) ReadPage(web string, title string) (*Page, error) {
	filename := pageToFilename(s.Root, web, title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body:body}, nil
}

func (s *FilePageStorage) WritePage(web string, p *Page) error {
	filename := pageToFilename(s.Root, web, p.Title)
	err := ioutil.WriteFile(filename, p.Body, 0644)
	if err != nil {
		return err
	}

	return commitPage(s, relativePathToPage(web, p.Title))
}

func (s *FilePageStorage) CreateWeb(web string) error {
	err := CopyDir(s.Root+"/_empty", s.Root+"/"+web)
	if err != nil {
		return err
	}

	return commitWeb(s, web)
}

func commitWeb(s *FilePageStorage, web string) error {
	if s.Repo != nil {
		sig := &git.Signature{
			Name:  "Guest User",
			Email: "guest@example.com",
			When:  time.Now(),
		}
		idx, err := s.Repo.Index()
		if err != nil {
			return err
		}
		err = idx.AddAll([]string{web}, git.IndexAddDefault, nil)
		if err != nil {
			return err
		}
		err = idx.Write()
		if err != nil {
			return err
		}
		treeId, err := idx.WriteTree()
		if err != nil {
			return err
		}
		currentBranch, err := s.Repo.Head()
		if err != nil {
			return err
		}
		currentTip, err := s.Repo.LookupCommit(currentBranch.Target())
		if err != nil {
			return err
		}
		tree, err := s.Repo.LookupTree(treeId)
		if err != nil {
			return err
		}
		message := web + " created"
		commitId, err := s.Repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
		if err != nil {
			return err
		}

		log.Info("Made commit " + commitId.String() + ".")

		go pushToOrigin(s)
	}
	return nil
}

func commitPage(s *FilePageStorage, path string) error {
	if s.Repo != nil {
		sig := &git.Signature{
			Name:  "Guest User",
			Email: "guest@example.com",
			When:  time.Now(),
		}
		idx, err := s.Repo.Index()
		if err != nil {
			return err
		}
		err = idx.AddByPath(path) //AddAll
		if err != nil {
			return err
		}
		err = idx.Write()
		if err != nil {
			return err
		}
		treeId, err := idx.WriteTree()
		if err != nil {
			return err
		}
		currentBranch, err := s.Repo.Head()
		if err != nil {
			return err
		}
		currentTip, err := s.Repo.LookupCommit(currentBranch.Target())
		if err != nil {
			return err
		}
		tree, err := s.Repo.LookupTree(treeId)
		if err != nil {
			return err
		}
		message := path + " updated"
		commitId, err := s.Repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
		if err != nil {
			return err
		}

		log.Info("Made commit " + commitId.String() + ".")

		go pushToOrigin(s)
	}
	return nil
}

func pullWikiData(repo *git.Repository, remote *git.Remote) {
	remoteCallbacks, err := getRemoteCallbacks()
	if err != nil {
		log.Warn(err)
		return
	}

	fetchOptions := &git.FetchOptions{RemoteCallbacks:*remoteCallbacks}

	if err := remote.Fetch([]string{}, fetchOptions, ""); err != nil {
		log.Warn(err)
		return
	}

	remoteBranch, err := repo.References.Lookup("refs/remotes/origin/master")
	if err != nil {
		log.Warn(err)
		return
	}

	remoteBranchID := remoteBranch.Target()
	annotatedCommit, err := repo.AnnotatedCommitFromRef(remoteBranch)
	if err != nil {
		log.Warn(err)
		return
	}
	mergeHeads := make([]*git.AnnotatedCommit, 1)
	mergeHeads[0] = annotatedCommit
	analysis, _, err := repo.MergeAnalysis(mergeHeads)
	if err != nil {
		log.Warn(err)
		return
	}

	// Get repo head
	head, err := repo.Head()
	if err != nil {
		log.Warn(err)
		return
	}

	if analysis & git.MergeAnalysisUpToDate != 0 {
		log.Info("Up to date with origin.")
		return
	}  else if analysis & git.MergeAnalysisNormal != 0 {
		// Just merge changes
		if err := repo.Merge([]*git.AnnotatedCommit{annotatedCommit}, nil, nil); err != nil {
			log.Warn(err)
			return
		}
		// Check for conflicts
		index, err := repo.Index()
		if err != nil {
			log.Warn(err)
			return
		}

		if index.HasConflicts() {
			log.Warn("Conflicts encountered. Please resolve them.")
			return
		}

		// Make the merge commit
		sig, err := repo.DefaultSignature()
		if err != nil {
			log.Warn(err)
			return
		}

		// Get Write Tree
		treeId, err := index.WriteTree()
		if err != nil {
			log.Warn(err)
			return
		}

		tree, err := repo.LookupTree(treeId)
		if err != nil {
			log.Warn(err)
			return
		}

		localCommit, err := repo.LookupCommit(head.Target())
		if err != nil {
			log.Warn(err)
			return
		}

		remoteCommit, err := repo.LookupCommit(remoteBranchID)
		if err != nil {
			log.Warn(err)
			return
		}

		repo.CreateCommit("HEAD", sig, sig, "", tree, localCommit, remoteCommit)

		// Clean up
		repo.StateCleanup()
		log.Info("Merged changes from remote origin.")
	} else if analysis & git.MergeAnalysisFastForward != 0 {
		// Fast-forward changes
		// Get remote tree
		remoteTree, err := repo.LookupTree(remoteBranchID)
		if err != nil {
			log.Warn(err)
			return
		}

		// Checkout
		if err := repo.CheckoutTree(remoteTree, nil); err != nil {
			log.Warn(err)
			return
		}

		branchRef, err := repo.References.Lookup("refs/heads/master")
		if err != nil {
			log.Warn(err)
			return
		}

		// Point branch to the object
		branchRef.SetTarget(remoteBranchID, "")
		if _, err := head.SetTarget(remoteBranchID, ""); err != nil {
			log.Warn(err)
			return
		}
		log.Info("Fast forward merged changes from remote origin.")
	} else {
		log.Warnf("Unexpected merge analysis result %d", analysis)
		return
	}

	return
}

func pushToOrigin(s *FilePageStorage) {
	if s.Origin != nil {
		err := s.Origin.Push([]string{"refs/heads/master"}, s.PushOptions)
		if err != nil {
			log.Error("Unable to push to origin: ", err)
		} else {
			log.Info("Pushed to origin.")
		}
	}
}

func initialiseWikiFromGitRepository(path string, initFromGitRepo string) error {
	opts := git.CloneOptions{
		Bare: false,
		RemoteCreateCallback: func(r *git.Repository, name, url string) (*git.Remote, git.ErrorCode) {
			remote, err := r.Remotes.Create("upstream", url)
			if err != nil {
				return nil, git.ErrGeneric
			}

			return remote, git.ErrOk
		},
	}

	_, err := git.Clone(initFromGitRepo, path, &opts)
	if err == nil {
		log.Info("Initialised new repository from '" + initFromGitRepo + "' at '" + path + "'.")
	}
	return err
}

func initialiseWikiAsCloneFromGitRepository(path string, cloneFromGitRepo string) error {
	remoteCallbacks, err := getRemoteCallbacks()
	if err != nil {
		return err
	}

	fetchOptions := &git.FetchOptions{RemoteCallbacks:*remoteCallbacks}

	opts := git.CloneOptions{
		Bare:           false,
		CheckoutBranch: "master",
		RemoteCreateCallback: func(r *git.Repository, name, url string) (*git.Remote, git.ErrorCode) {
			remote, err := r.Remotes.Create("origin", url)
			if err != nil {
				return nil, git.ErrGeneric
			}

			return remote, git.ErrOk
		},
		FetchOptions: fetchOptions,
	}

	log.Info("Cloning...")
	_, err = git.Clone(cloneFromGitRepo, path, &opts)
	if err == nil {
		log.Info("Cloned repository '" + cloneFromGitRepo + "' to '" + path + "'.")
	}
	return err
}

func configureOriginGitRepository(repo *git.Repository, originGitRepo string) error {
	if repo == nil {
		return errors.New("Unable to configure origin when data directory is not a git repository.")
	}

	origin, err := repo.Remotes.Create("origin", originGitRepo)
	if err != nil {
		return err
	}

	remoteCallbacks, err := getRemoteCallbacks()
	if err != nil {
		return err
	}

	pushOptions := &git.PushOptions{RemoteCallbacks:*remoteCallbacks}
	err = origin.Push([]string{"refs/heads/master"}, pushOptions)
	if err != nil {
		return err
	}

	return nil
}

func gitCredentials(username string, passphrase string, keyPath string) (func(string, string, git.CredType) (git.ErrorCode, *git.Cred), error) {
	errorCode, cred := git.NewCredSshKey(username, keyPath+".pub", keyPath, passphrase)

	if git.ErrorCode(errorCode) != git.ErrOk {
		return nil, errors.New("Invalid Credentials: " + string(errorCode))
	}
	return func(url string, username_from_url string, allowed_types git.CredType) (git.ErrorCode, *git.Cred) {
		return git.ErrorCode(errorCode), &cred
	}, nil
}

func gitCertCheck() (func(*git.Certificate, bool, string) git.ErrorCode) {
	return func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
		return git.ErrorCode(git.ErrOk)
	}
}

func checkForGitRepo(path string) (*git.Repository) {
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Warn("Data directory is not a git repository.")
		return nil
	}
	return repo
}

func getRemoteCallbacks() (*git.RemoteCallbacks, error) {
	credFunc, err := gitCredentials(os.Getenv("GOWIKI_GIT_USERNAME"), os.Getenv("GOWIKI_GIT_PASSPHRASE"), os.Getenv("GOWIKI_GIT_SSH_KEY_PATH"))
	if err != nil {
		return nil, err
	}

	certCheckFunc := gitCertCheck()
	remoteCallbacks := &git.RemoteCallbacks{
		CredentialsCallback:     credFunc,
		CertificateCheckCallback:certCheckFunc,
	}
	return remoteCallbacks, nil
}

func configureOrigin(repo *git.Repository) (*git.Remote, *git.PushOptions) {
	if repo == nil {
		return nil, nil
	}
	remote, err := repo.Remotes.Lookup("origin")
	if err != nil {
		log.Warn(err)
		log.Warn("No remote called origin configured to push to.")
		return nil, nil
	} else {
		remoteCallbacks, err := getRemoteCallbacks()
		if err == nil {
			pushOptions := &git.PushOptions{RemoteCallbacks:*remoteCallbacks}
			go pullWikiData(repo, remote);
			return remote, pushOptions
		} else {
			log.Warn("No GOWIKI_GIT credentials provided for Push")
			log.Warn(err)
			return remote, nil
		}
	}
}
