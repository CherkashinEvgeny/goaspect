package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	gofmt "go/format"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
)

var (
	dstPkgFlag     = flag.String("pkg", "", "Package name of generated code")
	dstPkgPathFlag = flag.String("path", "", "Package path of generated code")
	dstFileFlag    = flag.String("file", "", "Output file path")
)

func main() {
	flag.Usage = printUsage
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		printInvalidArgumentError("source package is missing")
		return
	}
	srcPkgArg := args[0]
	if srcPkgArg == "" {
		printInvalidArgumentError("source package is empty")
		return
	}

	srcPkg, err := parsePackage(srcPkgArg)
	if err != nil {
		printError("failed to parse package", err)
		return
	}
	var dstPkgPath string
	if *dstPkgPathFlag != "" {
		dstPkgPath = *dstPkgPathFlag
	} else if *dstFileFlag != "" {
		dstPkgDir, _ := path.Split(*dstFileFlag)
		dstPkgPath, err = resolvePackagePath(dstPkgDir)
		if err != nil {
			printWarning(fmt.Sprintf("failed to resolve destination package path, using '%s'", srcPkg.Path()), err)
			dstPkgPath = srcPkg.Path()
			err = nil
		}
	} else {
		dstPkgPath = srcPkg.Path()
	}
	var dstPkgName string
	if *dstPkgFlag != "" {
		dstPkgName = *dstPkgFlag
	} else {
		dstPkgName, err = resolvePackageName(dstPkgPath)
		if err != nil {
			printWarning(fmt.Sprintf("failed to resolve destination package name, using %s", srcPkg.Name()), err)
			dstPkgName = srcPkg.Name()
			err = nil
		}
	}
	options := parseAspectOptions(args[1:])
	aspects, err := findAspectsToGenerate(srcPkg, options)
	if err != nil {
		printError("failed to find aspects to generate", err)
		return
	}

	code, err := generate(config{
		DstPkgName:     dstPkgName,
		DstPackagePath: dstPkgPath,
		SrcPkg:         srcPkg,
		Aspects:        aspects,
	})
	if err != nil {
		printError("failed to generate code", err)
		return
	}
	formattedCode, err := format(code)
	if err != nil {
		printWarning("failed to format code", err)
	} else {
		code = formattedCode
	}

	var out io.Writer
	if *dstFileFlag == "" {
		out = os.Stdout
	} else {
		var file *os.File
		file, err := os.OpenFile(*dstFileFlag, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			printError("open file", err)
			return
		}
		defer func() {
			_ = file.Close()
		}()
		out = file
	}
	_, err = io.WriteString(out, code)
	if err != nil {
		printError("write code", err)
		return
	}
}

func parsePackage(path string) (*types.Package, error) {
	return importer.ForCompiler(token.NewFileSet(), "source", nil).Import(path)
}

func resolvePackagePath(dir string) (string, error) {
	stdout := bytes.NewBuffer(nil)
	stderr := bytes.NewBuffer(nil)
	cmd := exec.Command("go", "list", "-json", dir)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return "", errors.New(string(stderr.Bytes()))
		}
		return "", err
	}
	var stdoutJson struct {
		ImportPath string
	}
	err = json.Unmarshal(stdout.Bytes(), &stdoutJson)
	if err != nil {
		return "", err
	}
	return stdoutJson.ImportPath, nil
}

func resolvePackageName(path string) (string, error) {
	pkg, err := build.Import(path, ".", build.ImportComment)
	if err != nil {
		return "", err
	}
	return pkg.Name, nil
}

func parseAspectOptions(options []string) map[string]string {
	names := make(map[string]string, len(options))
	for _, option := range options {
		splited := strings.Split(option, "->")
		ifaceName := splited[0]
		var aspectName string
		if len(splited) == 2 {
			aspectName = splited[1]
		}
		names[ifaceName] = aspectName
	}
	return names
}

func findAspectsToGenerate(pkg *types.Package, options map[string]string) ([]aspectConfig, error) {
	ifaces := findNamedInterfaces(pkg)
	if len(options) == 0 {
		aspects := make([]aspectConfig, 0, len(ifaces))
		for ifaceName, iface := range ifaces {
			aspects = append(aspects, aspectConfig{
				IfaceName:  ifaceName,
				Iface:      iface,
				AspectName: ifaceName + "Aspect",
			})
		}
		return aspects, nil
	}
	aspects := make([]aspectConfig, 0, len(options))
	for ifaceName, aspectName := range options {
		iface, found := ifaces[ifaceName]
		if !found {
			return nil, errors.Errorf("interface='%s' not found", ifaceName)
		}
		if aspectName == "" {
			aspectName = ifaceName + "Aspect"
		}
		aspects = append(aspects, aspectConfig{
			IfaceName:  ifaceName,
			Iface:      iface,
			AspectName: aspectName,
		})
	}
	return aspects, nil
}

func findNamedInterfaces(pkg *types.Package) map[string]*types.Interface {
	items := map[string]*types.Interface{}
	pkgScope := pkg.Scope()
	names := pkgScope.Names()
	for _, name := range names {
		obj := pkgScope.Lookup(name)
		_, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		t := obj.Type()
		named, ok := t.(*types.Named)
		if !ok {
			continue
		}
		iface, ok := named.Underlying().(*types.Interface)
		if !ok {
			continue
		}
		items[name] = iface
	}
	return items
}

func format(code string) (string, error) {
	codeBytes, err := gofmt.Source([]byte(code))
	if err != nil {
		return "", err
	}
	return string(codeBytes), nil
}

const usage = `goaspect -pkg=[destination package name] -path=[destination package path] -file=[output file path] [source package] [interfaces]...
	[destination package name] - Package name of generated code. If empty, source package name will be used.
	[destination package path] - Package path of generated code. If empty, source package path will be used.
	[output file path]         - Path to output file. If empty, stdout will be used.
	[source package]           - Package path for which aspect code will be generated.
	[interfaces]               - List of interface names for which aspects will be generated. If empty, aspects will be generated for each interface in package.`

func printUsage() {
	fmt.Printf("%s\n", usage)
}

func printInvalidArgumentError(err string) {
	fmt.Printf("%s\n\n%s\n", err, usage)
}

func printWarning(description string, err error) {
	fmt.Printf("WARNING: %s\n\t%v", description, err)
}

func printError(description string, err error) {
	fmt.Printf("ERROR: %s\n\t%v", description, err)
}
