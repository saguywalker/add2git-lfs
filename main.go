package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	rice "github.com/GeertJohan/go.rice"

	"github.com/labstack/echo"
)

var remote string
var branch string
var uploadsDir string
var token string

func main() {
	flag.StringVar(&remote, "remote", "origin", "remote")
	flag.StringVar(&branch, "branch", "master", "branch")
	flag.StringVar(&uploadsDir, "folder", "sample-files", "folder to upload")
	flag.StringVar(&token, "token", "", "personal access token (https)")
	flag.Parse()

	os.MkdirAll(filepath.Join(".", uploadsDir), os.ModePerm)
	err := initLfs()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	assetHandler := http.FileServer(rice.MustFindBox("public").HTTPBox())
	e.GET("/", echo.WrapHandler(assetHandler))
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", assetHandler)))
	e.POST("/upload", handleUpload)
	e.POST("/pushfiles", handlePushFiles)
	go open("http://localhost:12358/")
	e.Logger.Fatal(e.Start(":12358"))
}

func handleUpload(c echo.Context) error {

	c.Request().ParseMultipartForm(32 << 20)
	form, err := c.MultipartForm()
	if err != nil {
		message := fmt.Sprintf("Error when parsing files %s", err.Error())
		return c.String(http.StatusBadRequest, message)
	}
	files := form.File["file"]

	var fullname string
	for _, file := range files {

		fullname = filepath.Join(".", uploadsDir, file.Filename)

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

func whichGit() (string, error) {
	out, err := exec.Command("which", "git").Output()
	if err != nil {
		return "", errors.New(string(out) + "\n" + err.Error())
	}
	return string(out), nil
}

func whichLfs() (string, error) {
	out, err := exec.Command("which", "git-lfs").Output()
	if err != nil {
		return "", errors.New(string(out) + "\n" + err.Error())
	}
	return string(out), nil
}

func initLfs() error {
	var err error
	out := make([]byte, 0)
	if runtime.GOOS == "windows" {
		initLfsCmd := fmt.Sprintf("git checkout -f && (git checkout %s || git checkout -b %s) && git lfs install && git lfs track \"%s/*\" && git add .gitattributes && git config http.sslVerify false", branch, branch, uploadsDir)
		out, err = exec.Command("cmd", "/C", initLfsCmd).Output()
	} else {
		exec.Command("git", "checkout -f").Output()

		_, err = exec.Command("git", "checkout", branch).Output()
		if err != nil {
			exec.Command("git", "checkout -b", branch)
		}

		out, err = exec.Command("git-lfs", "install").Output()
		if err != nil {
			return errors.New(string(out) + "\n" + err.Error())
		}

		out, err = exec.Command("git-lfs", "track", uploadsDir).Output()
		if err != nil {
			return errors.New(string(out) + "\n" + err.Error())
		}

		out, err = exec.Command("git", "add", ".gitattributes").Output()
		if err != nil {
			return errors.New(string(out) + "\n" + err.Error())
		}

		out, err = exec.Command("git", "config", "http.sslVerify", "false").Output()
	}

	if err != nil {
		return errors.New(string(out) + "\n" + err.Error())
	}

	return nil
}

func gitAddFile() error {
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

func gitCommitShell() error {
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

func gitPushShell() error {
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

func gitPushToken() error {
	var err error
	out := make([]byte, 0)
	gitURL := fmt.Sprintf("remote.%s.url", remote)

	if runtime.GOOS == "windows" {
		out, err = exec.Command("cmd", "/C", "git config "+gitURL).Output()
		if err != nil {
			return fmt.Errorf("Not found git url from git config %s", gitURL)
		}

		pushCommand := fmt.Sprintf("https://oauth2:%s@%s", token, string(out[8:len(out)-1]))
		out, err = exec.Command("cmd", "/C", "git push "+pushCommand).Output()
	} else {
		out, err = exec.Command("git", "config", gitURL).Output()
		if err != nil {
			return fmt.Errorf("Not found git url from git config %s", gitURL)
		}

		pushCommand := fmt.Sprintf("https://oauth2:%s@%s", token, string(out[8:len(out)-1]))
		out, err = exec.Command("git", "push", pushCommand).Output()
	}
	if err != nil {
		return errors.New(string(out) + "\n" + err.Error())
	}

	return nil
}

func handlePushFiles(c echo.Context) error {
	err := gitAddFile()
	if err != nil {
		errMsg := fmt.Sprintf("Error when running git add %s\n\n***************************************************\n%s", uploadsDir, err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	err = gitCommitShell()
	if err != nil {
		errMsg := fmt.Sprintf("Error when running git commit\n\n***************************************************\n%s", err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	if token == "" {
		err = gitPushShell()
	} else {
		err = gitPushShell()
	}
	err = gitPushShell()
	if err != nil {
		errMsg := fmt.Sprintf("Error when running git push\n\n***************************************************\n%s", err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	return c.Redirect(http.StatusMovedPermanently, "/")
}

func open(url string) error {
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
