package gitcommand

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/labstack/echo"
)

//Config is a bunch of configuration for a web application
type Config struct {
	Branch     string
	Email      string
	Remote     string
	Token      string
	UploadsDir string
	User       string
}

// NewConfig returns a new Config
func NewConfig(branch, email, remote, token, uploadsDir, user string) *Config {
	return &Config{
		Branch:     branch,
		Email:      email,
		Remote:     remote,
		UploadsDir: uploadsDir,
		User:       user,
	}
}

// InitLfs runs necessary commands before open a web application
// Including checkout to a specified branch, initialized git lfs, track a specified directory and add it to a worktree
func (config *Config) InitLfs() error {
	var err error
	out := make([]byte, 0)
	if runtime.GOOS == "windows" {
		initLfsCmd := fmt.Sprintf("git checkout -f && (git checkout %s || git checkout -b %s) && git lfs install && git lfs track \"%s/*\" && git add .gitattributes && git config http.sslVerify false", config.Branch, config.Branch, config.UploadsDir)
		out, err = exec.Command("cmd", "/C", initLfsCmd).Output()
	} else {
		exec.Command("git", "checkout -f").Output()

		err = exec.Command("git", "checkout", config.Branch).Run()
		if err != nil {
			exec.Command("git", "checkout", "-b", config.Branch).Run()
		}

		out, err = exec.Command("git-lfs", "install").Output()
		if err != nil {
			return fmt.Errorf("%s\n%s", string(out), err.Error())
		}

		out, err = exec.Command("git-lfs", "track", fmt.Sprintf("%s/*", config.UploadsDir)).Output()
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

// GitAddFile adds files in a specified directory to a worktree
func (config *Config) GitAddFile() error {
	var err error
	out := make([]byte, 0)
	if runtime.GOOS == "windows" {
		addCmd := fmt.Sprintf("git add %v", config.UploadsDir)
		out, err = exec.Command("cmd", "/C", addCmd).Output()
	} else {
		out, err = exec.Command("git", "add", config.UploadsDir).Output()
	}
	if err != nil {
		return fmt.Errorf("%s\n%s", string(out), err.Error())
	}

	return nil
}

// GitCommitFiles commits files according to a specified directory
func (config *Config) GitCommitFiles() error {
	var err error
	out := make([]byte, 0)
	if runtime.GOOS == "windows" {
		commitCmd := fmt.Sprintf("git commit -m upload-files-to-%s", config.UploadsDir)
		out, err = exec.Command("cmd", "/C", commitCmd).Output()
	} else {
		commitCmd := fmt.Sprintf("upload files to %s", config.UploadsDir)
		out, err = exec.Command("git", "commit", "-m", commitCmd).Output()
	}

	if err != nil {
		return fmt.Errorf("%s\n%s", string(out), err.Error())
	}

	return nil
}

// GitPushFiles pushs files to the specified remote and branch
func (config *Config) GitPushFiles() error {
	var err error
	out := make([]byte, 0)
	if runtime.GOOS == "windows" {
		gitPushCmd := fmt.Sprintf("git push %s %s", config.Remote, config.Branch)
		out, err = exec.Command("cmd", "/C", gitPushCmd).Output()
	} else {
		out, err = exec.Command("git", "push", config.Remote, config.Branch).Output()
	}
	if err != nil {
		return fmt.Errorf("%s\n%s", string(out), err.Error())
	}

	return nil
}

// GitPushToken pushs files to the specified remote and branch via a token.
func (config *Config) GitPushToken() error {
	var err error
	out := make([]byte, 0)
	gitURLCommand := fmt.Sprintf("remote.%s.url", config.Remote)

	if runtime.GOOS == "windows" {
		command := fmt.Sprintf("git config %s", gitURLCommand)
		out, err = exec.Command("cmd", "/C", command).Output()
	} else {
		out, err = exec.Command("git", "config", gitURLCommand).Output()

	}
	if err != nil {
		return fmt.Errorf("Not found git url from git config %s", gitURLCommand)
	}

	gitURL, isHTTPS, err := splitGitURL(out)
	if err != nil {
		return nil
	}

	var pushCommand string
	if isHTTPS {
		pushCommand = fmt.Sprintf("https://oauth2:%s@%s", config.Token, gitURL)
	} else {
		pushCommand = fmt.Sprintf("http://oauth2:%s@%s", config.Token, gitURL)
	}

	var command *exec.Cmd
	if runtime.GOOS == "windows" {
		runCommand := fmt.Sprintf("git push %s %s", pushCommand, config.Branch)
		command = exec.Command("cmd", "/C", runCommand)
		out, err = command.Output()
	} else {
		command = exec.Command("git", "push", pushCommand, config.Branch)
		out, err = command.Output()
	}

	if err != nil {
		return fmt.Errorf("%+v\n%s\n%s", command, string(out), err.Error())
	}

	return nil

}

// splitGitURL returns a GitURL for concatenating with token
func splitGitURL(url []byte) (string, bool, error) {
	if len(url) < 17 {
		return "", false, errors.New("too short url")
	}

	isHTTPS := false
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
		isHTTPS = true
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
		return "", false, errors.New("wrong format")
	}

	output := append(host, '/')
	output = append(output, user...)
	output = append(output, '/')
	output = append(output, repo...)

	if string(output[len(output)-4:]) != ".git" {
		repo = append(repo, []byte(".git")...)
	}

	return string(output), isHTTPS, nil
}

// ConfigUser configs the user.name and user.email if flags are provided
func (config *Config) ConfigUser(configType string) error {
	var configVar string

	switch strings.ToLower(configType) {
	case "name":
		configVar = config.User
		break
	case "email":
		configVar = config.Email
		break
	default:
		return errors.New("config type for commit should be either name or email")
	}

	out := make([]byte, 0)
	var err error

	if runtime.GOOS == "windows" {
		out, err = exec.Command("cmd", "/C", fmt.Sprintf("git config user.%s \"%s\"", configType, configVar)).Output()
	} else {
		out, err = exec.Command("git", "config", fmt.Sprintf("user.%s", configType), fmt.Sprintf("\"%s\"", configVar)).Output()
	}

	if err != nil {
		return fmt.Errorf("%s\n%s", out, err)
	}
	return nil
}

// HandleUpload handles the files uploading function
func (config *Config) HandleUpload(c echo.Context) error {

	c.Request().ParseMultipartForm(32 << 20)
	form, err := c.MultipartForm()
	if err != nil {
		message := fmt.Sprintf("Error when parsing files %s", err.Error())
		return c.String(http.StatusBadRequest, message)
	}
	files := form.File["file"]

	var fullname string
	for _, file := range files {

		fullname = filepath.Join(".", config.UploadsDir, file.Filename)

		src, err := file.Open()
		if err != nil {
			message := fmt.Sprintf("Error when opening %v", file.Filename)
			return c.String(http.StatusBadRequest, message)
		}
		defer src.Close()

		dst, err := os.Create(fullname)
		if err != nil {
			message := fmt.Sprintf("Error when opening %v", file.Filename)
			return c.String(http.StatusBadRequest, message)
		}
		defer dst.Close()

		if _, err = io.Copy(dst, src); err != nil {
			message := fmt.Sprintf("Error when opening %v", file.Filename)
			return c.String(http.StatusBadRequest, message)
		}
	}

	return c.String(http.StatusOK, "Files are uploaded")

}

// HandlePushFiles runs git add, commit and push
func (config *Config) HandlePushFiles(c echo.Context) error {
	err := config.GitAddFile()
	if err != nil {
		errMsg := fmt.Sprintf("Error when running git add %s\n\n***************************************************\n%s", config.UploadsDir, err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	err = config.GitCommitFiles()
	if err != nil {
		errMsg := fmt.Sprintf("Error when running git commit\n\n***************************************************\n%s", err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	if config.Token == "" {
		err = config.GitPushFiles()
	} else {
		err = config.GitPushToken()
	}
	if err != nil {
		errMsg := fmt.Sprintf("Error when running git push\n\n***************************************************\n%s", err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}
