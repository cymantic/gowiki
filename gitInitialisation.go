package main


import (
	"gopkg.in/libgit2/git2go.v25"
	log "github.com/Sirupsen/logrus"
	"errors"
	"os"
)

func gitCredentials(username string, passphrase string, keyPath string) (func(string, string, git.CredType) (git.ErrorCode, *git.Cred), error) {
	errorCode, cred := git.NewCredSshKey(username, keyPath+".pub", keyPath, passphrase)

	if git.ErrorCode(errorCode) != git.ErrOk {
		return nil, errors.New("Invalid Credentials: " + string(errorCode))
	}
	return func(url string, username_from_url string, allowed_types git.CredType) (git.ErrorCode, *git.Cred) {
		return git.ErrorCode(errorCode), &cred
	}, nil
}

func gitCertCheck() func(*git.Certificate, bool, string) git.ErrorCode {
	return func(cert *git.Certificate, valid bool, hostname string) git.ErrorCode {
		return git.ErrorCode(git.ErrOk)
	}
}

func getRemoteCallbacks() (*git.RemoteCallbacks, error) {
	credFunc, err := gitCredentials(os.Getenv("GOWIKI_GIT_USERNAME"), os.Getenv("GOWIKI_GIT_PASSPHRASE"), os.Getenv("GOWIKI_GIT_SSH_KEY_PATH"))
	if err != nil {
		return nil, err
	}

	certCheckFunc := gitCertCheck()
	remoteCallbacks := &git.RemoteCallbacks{
		CredentialsCallback:      credFunc,
		CertificateCheckCallback: certCheckFunc,
	}
	return remoteCallbacks, nil
}

func checkForGitRepo(path string) *git.Repository {
	repo, err := git.OpenRepository(path)
	if err != nil {
		log.Warn("Data directory is not a git repository.")
		return nil
	}
	return repo
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
			pushOptions := &git.PushOptions{RemoteCallbacks: *remoteCallbacks}
			go pullWikiData(repo, remote)
			return remote, pushOptions
		} else {
			log.Warn("No GOWIKI_GIT credentials provided for Push")
			log.Warn(err)
			return remote, nil
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

	fetchOptions := &git.FetchOptions{RemoteCallbacks: *remoteCallbacks}

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

	pushOptions := &git.PushOptions{RemoteCallbacks: *remoteCallbacks}
	err = origin.Push([]string{"refs/heads/master:refs/heads/master"}, pushOptions)
	if err != nil {
		return err
	}

	return nil
}

