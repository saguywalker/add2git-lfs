package gitcommand

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
)

//Config is a bunch of configuration for a web application
type Config struct {
	Branch     string
	Email      string
	OS         string
	Remote     string
	Token      string
	UploadsDir string
	User       string
}

// NewConfig returns a new Config
func NewConfig(branch, email, os, remote, token, uploadsDir, user string) *Config {
	return &Config{
		Branch:     branch,
		Email:      email,
		OS:         os,
		Remote:     remote,
		Token:      token,
		UploadsDir: uploadsDir,
		User:       user,
	}
}

// InitLfs runs necessary commands before open a web application
// Including checkout to a specified branch, initialized git lfs, track a specified directory and add it to a worktree
func (config *Config) InitLfs() error {
	var cmd string
	var args []string

	if config.OS == "windows" {
		cmd = "cmd"
		args = []string{"/C", fmt.Sprintf("git checkout -f && (git checkout %s || git checkout -b %s) && git lfs install && git lfs track \"%s/*\" && git add .gitattributes && git config http.sslVerify false", config.Branch, config.Branch, config.UploadsDir)}
	} else {
		cmd = "git"

		exec.Command(cmd, "checkout -f").Output()

		err := exec.Command(cmd, "checkout", config.Branch).Run()
		if err != nil {
			exec.Command(cmd, "checkout", "-b", config.Branch).Run()
		}

		out, err := exec.Command("git-lfs", "install").Output()
		if err != nil {
			return fmt.Errorf("%s\n%s", string(out), err.Error())
		}

		out, err = exec.Command("git-lfs", "track", fmt.Sprintf("%s/*", config.UploadsDir)).Output()
		if err != nil {
			return fmt.Errorf("%s\n%s", string(out), err.Error())
		}

		out, err = exec.Command(cmd, "add", ".gitattributes").Output()
		if err != nil {
			return fmt.Errorf("%s\n%s", string(out), err.Error())
		}

		args = []string{"config", "http.sslVerify", "false"}
	}

	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return fmt.Errorf("%s\n%s", string(out), err.Error())
	}

	return nil
}

// GitAddFile adds files in a specified directory to a worktree
func (config *Config) GitAddFile() error {
	var cmd string
	var args []string

	if config.OS == "windows" {
		cmd = "cmd"
		args = []string{"/C", fmt.Sprintf("git add %s", config.UploadsDir)}
	} else {
		cmd = "git"
		args = []string{"add", config.UploadsDir}
	}

	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return fmt.Errorf("%s\n%s", string(out), err.Error())
	}

	return nil
}

// GitCommitFiles commits files according to a specified directory
func (config *Config) GitCommitFiles() error {
	var cmd string
	var args []string

	if config.OS == "windows" {
		cmd = "cmd"
		args = []string{"/C", fmt.Sprintf("git commit -m upload-files-to%s", config.UploadsDir)}
	} else {
		cmd = "git"
		args = []string{"commit", "-m", fmt.Sprintf("upload files to %s", config.UploadsDir)}
	}

	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return fmt.Errorf("%s\n%s", string(out), err.Error())
	}

	return nil
}

// GitPushFiles pushs files to the specified remote and branch
func (config *Config) GitPushFiles() error {
	var cmd string
	var args []string

	if config.OS == "windows" {
		cmd = "cmd"
		args = []string{"/C", fmt.Sprintf("git push %s %s", config.Remote, config.Branch)}
	} else {
		cmd = "git"
		args = []string{"push", config.Remote, config.Branch}
	}

	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return fmt.Errorf("%s\n%s", string(out), err.Error())
	}

	return nil
}

// GitPushToken pushs files to the specified remote and branch via a token.
func (config *Config) GitPushToken() error {
	var cmd string
	var args []string

	gitURLCommand := fmt.Sprintf("remote.%s.url", config.Remote)

	if config.OS == "windows" {
		cmd = "cmd"
		args = []string{"/C", fmt.Sprintf("git config %s", gitURLCommand)}
	} else {
		cmd = "git"
		args = []string{"config", gitURLCommand}
	}

	out, err := exec.Command(cmd, args...).Output()
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

	if config.OS == "windows" {
		args = []string{"/C", fmt.Sprintf("git push %s %s", pushCommand, config.Branch)}
	} else {
		args = []string{"push", pushCommand, config.Branch}
	}

	out, err = exec.Command(cmd, args...).Output()
	if err != nil {
		return fmt.Errorf("%s\n%s", string(out), err.Error())
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
		output = append(output, []byte(".git")...)
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

	var cmd string
	var args []string

	if config.OS == "windows" {
		cmd = "cmd"
		args = []string{"/C", fmt.Sprintf("git config user.%s %s", configType, configVar)}
	} else {
		cmd = "git"
		args = []string{"config", fmt.Sprintf("user.%s", configType), fmt.Sprintf("\"%s\"", configVar)}
	}

	out, err := exec.Command(cmd, args...).Output()
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
