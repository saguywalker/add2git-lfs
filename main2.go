package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/kataras/iris"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var repo *git.Repository

const uploadsDir = "sample-files/"

func main() {
	app := iris.New()

	/*err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		if err.Error() != "already up-to-date" {
			panic(err)
		}
	}*/

	// Register templates
	app.RegisterView(iris.HTML("./views", ".html"))

	// Make the /public route path to statically serve the ./public/... contents
	app.HandleDir("/public", "./public")

	// Render the actual form
	// GET: http://localhost:12358
	app.Get("/", func(ctx iris.Context) {
		ctx.View("upload.html")
	})

	// Upload the file to the server
	// POST: http://localhost:12358/upload
	//app.Post("/upload", iris.LimitRequestBodySize(10<<20), handleUpload)
	app.Post("/upload", handleUpload2)

	// Start the server at http://localhost:12358
	app.Run(iris.Addr(":12358"))

	/*err := gitCommitShell()
	if err != nil{
		panic(err)
	}*/

	err := gitPushShell()
	if err != nil {
		panic(err)
	}

}

func handleUpload2(ctx iris.Context) {
	file, info, err := ctx.FormFile("file")
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.Application().Logger().Warnf("Error while uploading: %v", err.Error())
		return
	}

	defer file.Close()
	fname := info.Filename
	fullname := uploadsDir + fname
	// Create a file with the same name
	// assuming that you have a folder named 'uploads'
	out, err := os.OpenFile(fullname,
		os.O_WRONLY|os.O_CREATE, 0666)

	err = gitAddFile(fullname)
	if err != nil {
		panic(err)
	}

	err = gitCommitShell()
	if err != nil {
		panic(err)
	}

	/*err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
	})
	if err != nil {
		panic(err)
	}*/
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.Application().Logger().Warnf("Error while preparing the new file: %v", err.Error())
		return
	}
	defer out.Close()

	io.Copy(out, file)
}

func gitAddFile(filename string) error {
	addCmd := fmt.Sprintf("git add %v", filename)
	fmt.Println(addCmd)
	gitAddCmd := exec.Command("bash", "-c", "git add sample-files")
	_, err := gitAddCmd.Output()
	if err != nil {
		return err
	}
	return nil
}

func gitCommit(w *git.Worktree, commitMsg, name, email string) (plumbing.Hash, error) {
	commit, err := w.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  name,
			Email: email,
			When:  time.Now(),
		},
	})
	return commit, err
}

func gitCommitShell() error {
	gitCommitCmd := exec.Command("bash", "-c", "git commit -m \"upload sample files \"")
	out, err := gitCommitCmd.Output()
	if err != nil {
		return err
	}
	fmt.Println(string(out))

	return nil
}

func gitPushShell() error {
	gitPushCmd := exec.Command("bash", "-c", "git push origin master")
	out, err := gitPushCmd.Output()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
