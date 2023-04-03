package main

import (
	"flag"
	"fmt"
	"go/build"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"os"
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
	var dstPkg *build.Package
	if *dstPkgPathFlag != "" {
		dstPkg, err = findPackage(*dstPkgPathFlag)
	} else if *dstFileFlag != "" {
		dstPkgDir, _ := path.Split(*dstFileFlag)
		dstPkg, err = findPackageInDir(dstPkgDir)
	}
	// ignore destination package errors
	if err != nil {
		dstPkg = nil
		err = nil
	}
	var dstPkgName string
	if *dstPkgFlag != "" {
		dstPkgName = *dstPkgFlag
	} else if dstPkg != nil {
		dstPkgName = dstPkg.Name
	} else {
		dstPkgName = srcPkg.Name()
	}
	var dstPkgPath string
	if *dstPkgPathFlag != "" {
		dstPkgPath = *dstPkgPathFlag
	} else if dstPkg != nil {
		dstPkgPath = dstPkg.ImportPath
	} else {
		dstPkgPath = srcPkg.Path()
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

func findPackage(path string) (pkg *build.Package, err error) {
	return build.Import(path, ".", build.ImportComment)
}

func findPackageInDir(dir string) (pkg *build.Package, err error) {
	return build.ImportDir(dir, build.ImportComment)
}

func parsePackage(path string) (pkg *types.Package, err error) {
	return importer.ForCompiler(token.NewFileSet(), "source", nil).Import(path)
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

func printError(description string, err error) {
	fmt.Printf("%s:\n\t%v", description, err)
}
