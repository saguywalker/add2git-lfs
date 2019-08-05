package helper

import (
	"fmt"
	"os/exec"
	"runtime"
)

func WhichGit() (string, error) {
	out, err := exec.Command("which", "git").Output()
	if err != nil {
		return "", fmt.Errorf("%s\n%s", string(out), err.Error())
	}
	return string(out), nil
}

func WhichLfs() (string, error) {
	out, err := exec.Command("which", "git-lfs").Output()
	if err != nil {
		return "", fmt.Errorf("%s\n%s", string(out), err.Error())
	}
	return string(out), nil
}

func InitLfs(branch, uploadsDir string) error {
	var err error
	out := make([]byte, 0)
	if runtime.GOOS == "windows" {
		initLfsCmd := fmt.Sprintf("git checkout -f && (git checkout %s || git checkout -b %s) && git lfs install && git lfs track \"%s/*\" && git add .gitattributes && git config http.sslVerify false", branch, branch, uploadsDir)
		out, err = exec.Command("cmd", "/C", initLfsCmd).Output()
	} else {
		exec.Command("git", "checkout -f").Output()

		err = exec.Command("git", "checkout", branch).Run()
		if err != nil {
			exec.Command("git", "checkout", "-b", branch).Run()
		}

		out, err = exec.Command("git-lfs", "install").Output()
		if err != nil {
			return fmt.Errorf("%s\n%s", string(out), err.Error())
		}

		out, err = exec.Command("git-lfs", "track", uploadsDir).Output()
		if err != nil {
			return fmt.Errorf("%s\n%s", string(out), err.Error())
		}

		out, err = exec.Command("git", "add", ".gitattributes").Output()
		if err != nil {
			return fmt.Errorf("%s\n%s", string(out), err.Error())
		}

		out, err = exec.Command("git", "config", "http.sslVerify", "false").Output()
	}

	if err != nil {
		return fmt.Errorf("%s\n%s", string(out), err.Error())
	}

	return nil
}

func Open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
