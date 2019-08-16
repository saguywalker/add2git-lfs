package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/saguywalker/add2git-lfs/gitcommand"

	rice "github.com/GeertJohan/go.rice"

	"github.com/labstack/echo"
)

func main() {
	branch := flag.String("branch", "master", "branch")
	email := flag.String("email", "", "user.email for commit")
	//https := flag.Bool("https", true, "use https or not")
	port := flag.Int("port", 12358, "port for webapp")
	remote := flag.String("remote", "origin", "remote")
	token := flag.String("token", "", "personal access token")
	uploadsDir := flag.String("folder", "sample-files", "folder to upload files")
	user := flag.String("user", "", "user.name for commit")

	flag.Parse()

	var config *gitcommand.Config
	if runtime.GOOS == "windows" {
		config = gitcommand.NewConfig(*branch, *email, "windows", *remote, *token, *uploadsDir, *user)
	} else {
		config = gitcommand.NewConfig(*branch, *email, "linux", *remote, *token, *uploadsDir, *user)
	}

	if config.User != "" {
		err := config.ConfigUser("Name")
		if err != nil {
			panic(err)
		}
	}

	if *email != "" {
		err := config.ConfigUser("Email")
		if err != nil {
			panic(err)
		}
	}

	os.MkdirAll(filepath.Join(".", config.UploadsDir), os.ModePerm)
	err := config.InitLfs()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	//if *https == true {
	//		e.AutoTLSManager.HostPolicy = autocert.HostWhitelist(fmt.Sprintf("127.0.0.1:%d", *port))
	//		e.AutoTLSManager.Cache = autocert.DirCache("/var/www/.cache")
	//	}
	assetHandler := http.FileServer(rice.MustFindBox("public").HTTPBox())
	e.GET("/", echo.WrapHandler(assetHandler))
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", assetHandler)))
	e.POST("/upload", config.HandleUpload)
	e.POST("/pushfiles", config.HandlePushFiles)

	go Open(fmt.Sprintf("http://127.0.0.1:%d", *port))
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", *port)))
	//if *https == true {
	//		go Open(fmt.Sprintf("https://127.0.0.1:%d", *port))
	//		e.Logger.Fatal(e.StartAutoTLS(fmt.Sprintf(":%d", *port)))
	//	} else {
	//		go Open(fmt.Sprintf("http://127.0.0.1:%d", *port))
	//		e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", *port)))
	//	}
}

//Open a browser according to URL
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
