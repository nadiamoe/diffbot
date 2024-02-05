package gitea

import (
	"fmt"
	"log"
	"strconv"

	"code.gitea.io/sdk/gitea"
)

func PostOrUpdate(serverUrl, token, owner, repo, pr, commentBody string) error {
	prID, err := strconv.ParseInt(pr, 10, 64)
	if err != nil {
		return fmt.Errorf("parsing pr ID: %w", err)
	}

	log.Printf("Processing PR #%d", prID)

	client, err := gitea.NewClient(serverUrl, gitea.SetToken(token), gitea.SetUserAgent("diffbot"))
	if err != nil {
		return err
	}

	me, _, err := client.GetMyUserInfo()
	if err != nil {
		return fmt.Errorf("getting self info: %w", err)
	}

	log.Printf("Successfully logged into %s as %q", serverUrl, me.UserName)

	// giteaPR, _, err := client.GetPullRequest(owner, repo, prID)
	// if err != nil {
	// 	return fmt.Errorf("fetching pr details: %w", err)
	// }
	//
	// if giteaPR.State != gitea.StateOpen {
	// 	return fmt.Errorf("PR %d is %s, nothing to do", prID, giteaPR.State)
	// }

	comments, _, err := client.ListIssueComments(owner, repo, prID, gitea.ListIssueCommentOptions{})
	if err != nil {
		return fmt.Errorf("fetcihng pr comments: %w", err)
	}

	var ownComments []*gitea.Comment
	for _, comment := range comments {
		if comment.Poster.UserName == me.UserName {
			ownComments = append(ownComments, comment)
		}
	}

	switch len(ownComments) {
	case 0:
		log.Printf("No previous comments by %q found, creating one", me.UserName)
		_, _, err := client.CreateIssueComment(owner, repo, prID, gitea.CreateIssueCommentOption{
			Body: commentBody,
		})
		if err != nil {
			return fmt.Errorf("creating comment: %w", err)
		}

	case 2:
		log.Printf(
			"Warning: More than one comment by myself detected. I am not smart enough to handle this, will edit my" +
				" last comment only. Sorry about that.",
		)
		fallthrough

	case 1:
		log.Printf("Found my previous comment, updating it")

		_, _, err := client.EditIssueComment(owner, repo, ownComments[0].ID, gitea.EditIssueCommentOption{
			Body: commentBody,
		})
		if err != nil {
			return fmt.Errorf("editing comment: %w", err)
		}

	}

	return nil
}
