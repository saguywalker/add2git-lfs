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
	"time"

	"github.com/labstack/echo"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

//Config is a bunch of configuration for a web application
type Config struct {
	Repo       *git.Repository
	Branch     string
	Email      string
	Remote     string
	Token      string
	UploadsDir string
	User       string
	Os         string
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
func InitLfs(branch, email, remote, token, uploadsDir, user string) (*Config, error) {
	path, _ := os.Getwd()
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	config := &Config{
		Repo:       repo,
		Branch:     branch,
		Email:      email,
		Remote:     remote,
		UploadsDir: uploadsDir,
		User:       user,
	}

	if config.User == "" {
		err = config.ConfigUser("name")
		if err != nil {
			return nil, err
		}
	}

	if config.Email == "" {
		err = config.ConfigUser("email")
		if err != nil {
			return nil, err
		}
	}

	headRef, err := config.Repo.Head()
	if err != nil {
		return nil, err
	}

	branchRef := plumbing.NewBranchReferenceName(config.Branch)
	ref := plumbing.NewHashReference(branchRef, headRef.Hash())

	err = config.Repo.Storer.SetReference(ref)
	if err != nil {
		return nil, err
	}

	worktree, err := config.Repo.Worktree()
	if err != nil {
		return nil, err
	}

	err = worktree.Checkout(&git.CheckoutOptions{
		Force:  true,
		Branch: branchRef,
	})
	if err != nil {
		return nil, err
	}

	var cmd string
	var args []string
	if runtime.GOOS == "windows" {
		config.Os = "windows"

		cmd = "cmd"
		args = []string{"/C", fmt.Sprintf("git lfs install && git lfs track %s/*", config.UploadsDir)}
	} else {
		config.Os = "linux"

		cmd = "git-lfs"
		out, err := exec.Command(cmd, "install").Output()
		if err != nil {
			return nil, fmt.Errorf("%s\n%s", string(out), err.Error())
		}

		args = []string{"track", fmt.Sprintf("%s/*", config.UploadsDir)}
	}

	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return nil, fmt.Errorf("%s\n%s", string(out), err.Error())
	}

	err = config.GitAddFile(".gitattributes")
	if err != nil {
		return nil, err
	}

	return config, nil
}

// GitAddFile adds files in a specified directory to a worktree
func (config *Config) GitAddFile(filename string) error {
	var cmd string
	var args []string

	if config.Os == "windows" {
		cmd = "cmd"
		args = []string{"/C", fmt.Sprintf("git add %v", filename)}
	} else {
		cmd = "git"
		args = []string{"add", filename}
	}

	out, err := exec.Command(cmd, args...).Output()

	if err != nil {
		return fmt.Errorf("%s\n%s", string(out), err.Error())
	}

	return nil
}

// GitCommitFiles commits files according to a specified directory
func (config *Config) GitCommitFiles() error {
	worktree, err := config.Repo.Worktree()
	if err != nil {
		return err
	}

	_, err = worktree.Commit(fmt.Sprintf("upload files to %s", config.UploadsDir), &git.CommitOptions{
		Author: &object.Signature{
			Name:  config.User,
			Email: config.Email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// GitPushFiles pushs files to the specified remote and branch
func (config *Config) GitPushFiles() error {
	err := config.Repo.Push(&git.PushOptions{
		RemoteName: config.Remote,
	})

	return err
}

// GitPushToken pushs files to the specified remote and branch via a token.
func (config *Config) GitPushToken() error {
	var err error
	out := make([]byte, 0)
	gitURLCommand := fmt.Sprintf("remote.%s.url", config.Remote)

	if config.Os == "windows" {
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
	if config.Os == "windows" {
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

// ConfigUser configs the user.name and user.email if flags are not provided
func (config *Config) ConfigUser(configType string) error {
	var cmd string
	var args []string

	if config.Os == "windows" {
		cmd = "cmd"
		args = []string{"/C", fmt.Sprintf("git config user.%s", configType)}
	} else {
		cmd = "git"
		args = []string{"config", fmt.Sprintf("user.%s", configType)}
	}

	out, err := exec.Command(cmd, args...).Output()

	if err != nil {
		return fmt.Errorf("%s\n%s", out, err)
	}

	switch strings.ToLower(configType) {
	case "name":
		config.User = string(out[:len(out)-1])
		break
	case "email":
		config.Email = string(out[:len(out)-1])
		break
	default:
		return errors.New("config type for commit should be either name or email")
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
	err := config.GitAddFile(fmt.Sprintf("%s/*", config.UploadsDir))
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
