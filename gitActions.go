package main

import (
	"gopkg.in/libgit2/git2go.v25"
	"time"
	log "github.com/Sirupsen/logrus"
)

type GitWork struct {
	Action func()
}

var GitWorkQueue = make(chan GitWork, 10)

func startGitWorker() {
	go func() {
		for {
			work := <-GitWorkQueue
			work.Action()
		}
	}()
}

func pushToOrigin(r *FileWikiRepository) {
	log.Info("Looking up origin")
	origin, err := r.Repo.Remotes.Lookup("origin")
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Going to push to origin")
	err = origin.Push([]string{"refs/heads/master:refs/heads/master"}, r.PushOptions)
	if err != nil {
		log.Error("Unable to push to origin: ", err)
		return
	} else {
		log.Info("Pushed to origin.")
	}
}

func commitWeb(r *FileWikiRepository, web string) {
	if r.Repo != nil {
		sig := &git.Signature{
			Name:  "Guest User",
			Email: "guest@example.com",
			When:  time.Now(),
		}
		idx, err := r.Repo.Index()
		if err != nil {
			log.Error(err)
			return
		}
		err = idx.AddAll([]string{web}, git.IndexAddDefault, nil)
		if err != nil {
			log.Error(err)
			return
		}
		err = idx.Write()
		if err != nil {
			log.Error(err)
			return
		}
		treeId, err := idx.WriteTree()
		if err != nil {
			log.Error(err)
			return
		}
		currentBranch, err := r.Repo.Head()
		if err != nil {
			log.Error(err)
			return
		}
		currentTip, err := r.Repo.LookupCommit(currentBranch.Target())
		if err != nil {
			log.Error(err)
			return
		}
		tree, err := r.Repo.LookupTree(treeId)
		if err != nil {
			log.Error(err)
			return
		}
		message := web + " created"
		commitId, err := r.Repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
		if err != nil {
			log.Error(err)
			return
		}

		log.Info("Made commit " + commitId.String() + ".")

		pushToOrigin(r)
	}
	return
}

func commitPage(r *FileWikiRepository, path string) {
	if r.Repo != nil {
		sig := &git.Signature{
			Name:  "Guest User",
			Email: "guest@example.com",
			When:  time.Now(),
		}
		idx, err := r.Repo.Index()
		if err != nil {
			log.Error(err)
			return
		}
		err = idx.AddByPath(path) //AddAll
		if err != nil {
			log.Error(err)
			return
		}
		err = idx.Write()
		if err != nil {
			log.Error(err)
			return
		}
		treeId, err := idx.WriteTree()
		if err != nil {
			log.Error(err)
			return
		}
		currentBranch, err := r.Repo.Head()
		if err != nil {
			log.Error(err)
			return
		}
		currentTip, err := r.Repo.LookupCommit(currentBranch.Target())
		if err != nil {
			log.Error(err)
			return
		}
		tree, err := r.Repo.LookupTree(treeId)
		if err != nil {
			log.Error(err)
			return
		}
		message := path + " updated"
		commitId, err := r.Repo.CreateCommit("HEAD", sig, sig, message, tree, currentTip)
		if err != nil {
			log.Error(err)
			return
		}

		log.Info("Made commit " + commitId.String() + ".")

		pushToOrigin(r)
	}
	return
}

func pullWikiData(repo *git.Repository, remote *git.Remote) {
	remoteCallbacks, err := getRemoteCallbacks()
	if err != nil {
		log.Warn(err)
		return
	}

	fetchOptions := &git.FetchOptions{RemoteCallbacks: *remoteCallbacks}

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

	if analysis&git.MergeAnalysisUpToDate != 0 {
		log.Info("Up to date with origin.")
		return
	} else if analysis&git.MergeAnalysisNormal != 0 {
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
	} else if analysis&git.MergeAnalysisFastForward != 0 {
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
