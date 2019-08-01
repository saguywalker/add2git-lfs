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

func main() {
	flag.StringVar(&remote, "remote", "origin", "remote")
	flag.StringVar(&branch, "branch", "master", "branch")
	flag.StringVar(&uploadsDir, "folder", "sample-files", "folder to upload")
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

func initLfs() error {
	var err error
	out := make([]byte, 0)
	initLfsCmd := fmt.Sprintf("git checkout -f && (git checkout %s || git checkout -b %s) && git lfs install && git lfs track \"%s/*\" && git add .gitattributes && git config http.sslVerify false", branch, branch, uploadsDir)
	if runtime.GOOS == "windows" {
		out, err = exec.Command("cmd", "/C", initLfsCmd).Output()
	} else {
		out, err = exec.Command("sh", "-c", initLfsCmd).Output()
	}
	if err != nil {
		return errors.New(string(out))
	}

	return nil
}

func gitAddFile(filename string) error {
	var err error
	out := make([]byte, 0)
	addCmd := fmt.Sprintf("git add %v", filename)
	if runtime.GOOS == "windows" {
		out, err = exec.Command("cmd", "/C", addCmd).Output()
	} else {
		out, err = exec.Command("sh", "-c", addCmd).Output()
	}
	if err != nil {
		return errors.New(string(out))
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
		commitCmd := fmt.Sprintf("git commit -m \"upload files to %s\"", uploadsDir)
		out, err = exec.Command("sh", "-c", commitCmd).Output()
	}

	if err != nil {
		return errors.New(string(out))
	}

	return nil
}

func gitPushShell(remote, branch string) error {
	var err error
	out := make([]byte, 0)
	gitPushCmd := fmt.Sprintf("git push %s %s", remote, branch)
	if runtime.GOOS == "windows" {
		out, err = exec.Command("cmd", "/C", gitPushCmd).Output()
	} else {
		out, err = exec.Command("sh", "-c", gitPushCmd).Output()
	}
	if err != nil {
		return errors.New(string(out))
	}

	return nil
}

func handlePushFiles(c echo.Context) error {
	err := gitAddFile(uploadsDir)
	if err != nil {
		errMsg := fmt.Sprintf("Error when running git add %s\n\n***************************************************\n%s", uploadsDir, err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	err = gitCommitShell()
	if err != nil {
		errMsg := fmt.Sprintf("Error when running git commit\n\n***************************************************\n%s", err.Error())
		return c.String(http.StatusExpectationFailed, errMsg)
	}

	err = gitPushShell(remote, branch)
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
