package main

import (
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"go/importer"
	"go/token"
	"go/types"
	"io"
	"os"
	"strings"
)

var (
	outPkgFlag = flag.String("pkg", "", "Output package")
	outFlag    = flag.String("out", "", "Output file name")
)

func main() {
	flag.Parse()
	flag.Usage = printUsage
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
	srcPkg, err := parseSrcPackage(srcPkgArg)
	if err != nil {
		printError("failed to parse package", err)
		return
	}
	options := parseAspectOptions(args[1:])
	aspects, err := findAspectsToGenerate(srcPkg, options)
	if err != nil {
		printError("failed to find aspects to generate", err)
		return
	}
	dstPkgName := srcPkg.Name()
	if *outPkgFlag != "" {
		dstPkgName = *outPkgFlag
	}
	code, err := generate(config{
		DstPkgName: dstPkgName,
		SrcPkg:     srcPkg,
		Aspects:    aspects,
	})
	if outFlag == nil {
		fmt.Println(code)
		return
	}
	var out io.Writer
	if *outFlag == "" {
		out = os.Stdout
	} else {
		var file *os.File
		file, err := os.OpenFile(*outFlag, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
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

func parseSrcPackage(name string) (pkg *types.Package, err error) {
	return importer.ForCompiler(token.NewFileSet(), "source", nil).Import(name)
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

const usage = `goaspect -pkg=[destination package name] -out=[output file name] [source package] [interfaces]...
	[destination package name] - Package name of generated code. If empty, source package name will be used.
	[output file name]         - Path to output file. If empty, stdout will be used.
	[source package]           - Golang package path for which aspect code will be generated.
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
