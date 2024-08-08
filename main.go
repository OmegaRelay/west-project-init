package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path"
)

const kHeader = "================================================================================="

var (
	mkdirPerms os.FileMode = 0777
	touchPerms os.FileMode = 0666

	projectPath string = ""
)

//go:embed template
var templateFs embed.FS

func main() {

	projectPath = os.Args[1]

	fmt.Printf("\n\n%s\n\n", kHeader)
	os.Mkdir(projectPath, mkdirPerms)
	os.Chdir(projectPath)

	os.Mkdir("app", mkdirPerms)
	os.Mkdir("app/src", mkdirPerms)

	os.Create("app/prj.conf")

	// filepath.Walk("template")
	i, _ := templateFs.ReadDir("template")
	for _, file := range i {
		if file.Type().IsRegular() {
			copyTemplateFile(file)
		}
	}

	runCmd("git", "init")
	runCmd("python3", "-m", "venv", ".venv")
	runCmd("source", ".venv/bin/activate")
	runCmd("python3", "-m", "pip", "install", "west")

	fmt.Printf("\n\n%s\n\n", kHeader)
	fmt.Println("Project setup complete!")
	fmt.Println("Now setup app/west.yml, run make bootstrap, and make something cool with Zephyr")
}

// Keywords within @@ symbols are replaced with dynamic components
func replaceKeyWords(b []byte) (ret []byte, err error) {
	array := bytes.Split(b, []byte("@"))

	for index, val := range array {
		if index%2 != 1 {
			continue
		}
		switch string(val) {
		case "PROJECT_NAME":
			array[index] = []byte(path.Base(projectPath))
		default:
			continue
		}
	}
	ret = bytes.Join(array, nil)
	return
}

// Copy template files from template dir to project dir
func copyTemplateFile(file os.DirEntry) {
	content, _ := templateFs.ReadFile("template/" + file.Name())
	content, _ = replaceKeyWords(content)

	filePath := bytes.ReplaceAll([]byte(file.Name()), []byte(".template"), []byte(""))
	filePath = bytes.ReplaceAll(filePath, []byte("DOT_"), []byte("."))
	filePath = bytes.ReplaceAll(filePath, []byte("@"), []byte("/"))

	os.WriteFile(string(filePath), content, touchPerms)
}

// Wrapper around exec.Command to start, attach and print output of the command
func runCmd(command string, arg ...string) error {
	cmd := exec.Command(command, arg...)

	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	return nil
}
