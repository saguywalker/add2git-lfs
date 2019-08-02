package gitcommand

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
)

func GitAddFile(uploadsDir string) error {
	var err error
	out := make([]byte, 0)
	if runtime.GOOS == "windows" {
		addCmd := fmt.Sprintf("git add %v", uploadsDir)
		out, err = exec.Command("cmd", "/C", addCmd).Output()
	} else {
		out, err = exec.Command("git", "add", uploadsDir).Output()
	}
	if err != nil {
		return errors.New(string(out) + "\n" + err.Error())
	}

	return nil
}

func GitCommitShell(uploadsDir string) error {
	var err error
	out := make([]byte, 0)
	if runtime.GOOS == "windows" {
		commitCmd := fmt.Sprintf("git commit -m upload-files-to-%s", uploadsDir)
		out, err = exec.Command("cmd", "/C", commitCmd).Output()
	} else {
		commitCmd := fmt.Sprintf("\"upload files to %s\"", uploadsDir)
		out, err = exec.Command("git", "commit", "-m", commitCmd).Output()
	}

	if err != nil {
		return errors.New(string(out) + "\n" + err.Error())
	}

	return nil
}

func GitPushShell(remote, branch string) error {
	var err error
	out := make([]byte, 0)
	if runtime.GOOS == "windows" {
		gitPushCmd := fmt.Sprintf("git push %s %s", remote, branch)
		out, err = exec.Command("cmd", "/C", gitPushCmd).Output()
	} else {
		out, err = exec.Command("git", "push", remote, branch).Output()
	}
	if err != nil {
		return errors.New(string(out) + "\n" + err.Error())
	}

	return nil
}

func GitPushToken(remote, branch, token string) error {
	var err error
	out := make([]byte, 0)
	gitURL := fmt.Sprintf("remote.%s.url", remote)

	if runtime.GOOS == "windows" {
		out, err = exec.Command("cmd", "/C", "git config "+gitURL).Output()
		if err != nil {
			return fmt.Errorf("Not found git url from git config %s", gitURL)
		}

		pushCommand := fmt.Sprintf("https://oauth2:%s@%s", token, string(out[8:len(out)-1]))
		out, err = exec.Command("cmd", "/C", "git push "+pushCommand+" "+branch).Output()
	} else {
		out, err = exec.Command("git", "config", gitURL).Output()
		if err != nil {
			return fmt.Errorf("Not found git url from git config %s", gitURL)
		}

		pushCommand := fmt.Sprintf("https://oauth2:%s@%s", token, string(out[8:len(out)-1]))
		out, err = exec.Command("git", "push", pushCommand, branch).Output()
	}
	if err != nil {
		return errors.New(string(out) + "\n" + err.Error())
	}

	return nil
}
