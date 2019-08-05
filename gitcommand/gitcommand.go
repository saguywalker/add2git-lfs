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
	gitURLCommand := fmt.Sprintf("remote.%s.url", remote)

	if runtime.GOOS == "windows" {
		out, err = exec.Command("cmd", "/C", "git config "+gitURLCommand).Output()
		if err != nil {
			return fmt.Errorf("Not found git url from git config %s", gitURLCommand)
		}

		gitURL, err := splitGitURL(out)
		if err != nil {
			return nil
		}

		pushCommand := fmt.Sprintf("https://oauth2:%s@%s", token, gitURL)
		out, err = exec.Command("cmd", "/C", "git push "+pushCommand+" "+branch).Output()
	} else {
		out, err = exec.Command("git", "config", gitURLCommand).Output()
		if err != nil {
			return fmt.Errorf("Not found git url from git config %s", gitURLCommand)
		}

		gitURL, err := splitGitURL(out)
		if err != nil {
			return nil
		}

		pushCommand := fmt.Sprintf("https://oauth2:%s@%s", token, gitURL)
		out, err = exec.Command("git", "push", pushCommand, branch).Output()
	}

	if err != nil {
		return errors.New(string(out) + "\n" + err.Error())
	}

	return nil
}

func splitGitURL(url []byte) (string, error) {
	if len(url) < 17 {
		return "", errors.New("too short url")
	}

	var host []byte
	var user []byte
	var repo []byte
	var endHost int
	var endUser int

	if string(url[:8]) == "https://" {
		url = url[8:]
		for i, x := range url {
			if x == '/' && endHost == 0 {
				endHost = i
				host = url[:endHost]
			} else if x == '/' && endHost != 0 {
				endUser = i
				user = url[endHost+1 : endUser]

				repo = url[endUser+1:]
			}
		}
	} else if string(url[:7]) == "http://" {
		url = url[7:]
		for i, x := range url {
			if x == '/' && endHost == 0 {
				endHost = i
				host = url[:endHost]
			} else if x == '/' && endHost != 0 {
				endUser = i
				user = url[endHost+1 : endUser]

				repo = url[endUser+1:]
			}
		}
	} else if string(url[:4]) == "git@" {
		url = url[4:]
		for i, x := range url {
			if x == ':' {
				endHost = i
				host = url[:endHost]
			} else if x == '/' {
				endUser = i
				user = url[endHost+1 : endUser]

				repo = url[endUser+1:]
			}
		}
	} else {
		return "", errors.New("wrong format")
	}

	output := append(host, '/')
	output = append(output, user...)
	output = append(output, '/')
	output = append(output, repo...)

	if string(output[len(output)-4:]) != ".git" {
		repo = append(repo, []byte(".git")...)
	}

	return string(output), nil
}
