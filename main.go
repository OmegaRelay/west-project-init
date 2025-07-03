package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"syscall"

	flag "github.com/spf13/pflag"
)

const kHeader = `  
    \\                //   ========     =====     ==============
     \\              //   ||          ||     \\         ||
      \\            //    ||          \\                ||
       \\    __    //     ||======      =====           ||
        \\  //\\  //      ||                 \\         ||
         \\//  \\//       ||          \\     ||         ||
          \/    \/         ========     =====           ||
`

const kDivider = "================================================================================="

var (
	mkdirPerms os.FileMode = 0777
	touchPerms os.FileMode = 0664

	projectPath string = ""
)

//go:embed template
var templateFs embed.FS

//go:embed VERSION
var version string

var pythonRequirements = []string{
	"west", "pyelftools", "intelhex", "pyserial",
}

func main() {
	parseFlags()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("Error, must provide project name")
		return
	}

	// the first positional argument is the directory to initialise
	projectPath := args[0]

	initDir(projectPath)
}

func initDir(dirPath string) {
	if dirPath == "" {
		fmt.Println("Error: must provide project name")
		return
	}

	fmt.Printf("\n\n%s\n", kDivider)
	fmt.Printf("%s", kHeader)
	fmt.Printf("\n\n%s\n\n", kDivider)
	err := os.Mkdir(dirPath, mkdirPerms)
	if err != nil {
		if e, ok := err.(*os.PathError); ok && e.Err != syscall.EEXIST {
			fmt.Println(err)
			return
		}
	}

	err = os.Chdir(dirPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	templateContents, err := templateFs.ReadDir("template")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Copying template directory...\n")
	copyTemplateContents("", templateContents)
	fmt.Printf("\n\n")

	err = runCmd("git", "init")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = runCmd("python3", "-m", "venv", ".venv")
	if err != nil {
		fmt.Println(err)
		return
	}

	pipExe := ".venv/bin/pip"
	for _, requirement := range pythonRequirements {
		err = runCmd(pipExe, "install", requirement)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	err = runCmd(".venv/bin/west", "init", "-l", "zephyr")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("\n\n%s\n\n", kDivider)
	fmt.Printf("Project setup complete!\n\n")
	fmt.Printf("Add required third party projects through zephyr/west.yml\n")
	fmt.Printf("Run `source .venv/bin/activate` to set the project's virtual environment\n")
	fmt.Printf("Run `west update` to update the project\n")
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

func copyTemplateContents(path string, entries []fs.DirEntry) {
	for _, entry := range entries {
		if entry.Type().IsDir() {
			var dirPath string
			if path == "" {
				dirPath = entry.Name()
			} else {
				dirPath = path + "/" + entry.Name()
			}
			fmt.Printf("%s\n", dirPath)
			os.Mkdir(dirPath, mkdirPerms)
			entryContents, err := templateFs.ReadDir("template/" + dirPath)
			if err != nil {
				panic(err)
			}
			copyTemplateContents(dirPath, entryContents)
		} else if entry.Type().IsRegular() {
			var filePath string
			if path == "" {
				filePath = entry.Name()
			} else {
				filePath = path + "/" + entry.Name()
			}
			fmt.Printf("%s\n", filePath)
			content, err := templateFs.ReadFile("template/" + filePath)
			if err != nil {
				panic(err)
			}
			content, err = replaceKeyWords(content)
			if err != nil {
				panic(err)
			}
			err = os.WriteFile(filePath, content, touchPerms)
			if err != nil {
				panic(err)
			}
		}
	}
}

// Wrapper around exec.Command to start, attach and print output of the command
func runCmd(command string, arg ...string) error {
	cmd := exec.Command(command, arg...)

	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		return err
	}

	cmd.Start()
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Println(string(tmp))
		if err != nil {
			break
		}
	}
	return nil
}

func parseFlags() {
	versionFlag := flag.BoolP("version", "V", false, "Print the version")
	helpFlag := flag.BoolP("help", "h", false, "Print this help message")
	flag.Parse()

	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Println("v" + version)
		os.Exit(0)
	}

}

func printHelp() {
	fmt.Printf(`
Create and Set up a new West project in the given directory. 
Will initialise a git directory with a python virtual environment and makefile 
for building with the Zephyr RTOS using West.

Once created, the new directory has a makefile where "make bootstrap" can be 
used to clone the Zephyr project into the directory under the "third-party" 
directory.

Alternatively, "source .venv/bin/activate" can be used to gain access a python 
virtual environment with the west tool.


Usage:
	%s [flags] <project-path>"
	
Flags:
`, os.Args[0])

	flag.PrintDefaults()
}
