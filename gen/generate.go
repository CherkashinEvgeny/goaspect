package main

import (
	"go/types"

	. "github.com/CherkashinEvgeny/gogen"
	tgen "github.com/CherkashinEvgeny/gogen/types"
	"github.com/CherkashinEvgeny/gogen/utils"
)

const copyright = "Code generated by github.com/CherkashinEvgeny/goaspect. DO NOT EDIT."

type config struct {
	DstPkgName     string
	DstPackagePath string
	SrcPkg         *types.Package
	ImportSrcPkg   bool
	Aspects        []aspectConfig
}

type aspectConfig struct {
	IfaceName  string
	Iface      *types.Interface
	AspectName string
}

func generate(cfg config) (code string, err error) {
	imports := Imports(
		SmartImport("reflect", "", "reflect"),
		SmartImport("aspect", "", "github.com/CherkashinEvgeny/goaspect"),
	)
	if cfg.DstPackagePath != "" {
		imports.Add(SmartImport("", "", cfg.DstPackagePath))
	}
	aspects := Blocks()
	for _, aspectCfg := range cfg.Aspects {
		aspects.Add(generateInterfaceAspect(cfg.SrcPkg, aspectCfg))
	}
	pkg := Pkg(copyright, cfg.DstPkgName, imports, aspects)
	return Stringify(pkg), nil
}

func generateInterfaceAspect(pkg *types.Package, aspectCfg aspectConfig) Code {
	reflectTypeId := utils.Private(aspectCfg.IfaceName + "Type")
	return Blocks(
		Assign(
			Var(Id(reflectTypeId), nil),
			Join(Raw("reflect.TypeOf((*"), SmartQual(pkg.Name(), pkg.Path(), aspectCfg.IfaceName), Raw(")(nil)).Elem()")),
		),
		Type(aspectCfg.AspectName, Struct(FieldDecls(
			FieldDecl("Impl", SmartQual(pkg.Name(), pkg.Path(), aspectCfg.IfaceName)),
			FieldDecl("Factory", SmartQual("aspect", "github.com/CherkashinEvgeny/goaspect", "Factory")),
		))),
		generateAspectMethods(reflectTypeId, aspectCfg.AspectName, aspectCfg.Iface),
	)
}

func generateAspectMethods(reflectTypeId string, aspectName string, iface *types.Interface) Code {
	methods := make([]Code, 0)
	n := iface.NumEmbeddeds()
	for i := 0; i < n; i++ {
		embedded := iface.EmbeddedType(i)
		underlying := embedded.Underlying()
		embeddedIface, ok := underlying.(*types.Interface)
		if !ok {
			continue
		}
		methods = append(methods, generateAspectMethods(reflectTypeId, aspectName, embeddedIface))
	}
	n = iface.NumMethods()
	for i := 0; i < n; i++ {
		method := iface.Method(i)
		methodType := method.Type()
		sign, ok := methodType.(*types.Signature)
		if !ok {
			continue
		}
		methods = append(methods, generateAspectMethod(reflectTypeId, aspectName, method.Name(), sign))
	}
	return Blocks(methods...)
}

func generateAspectMethod(reflectTypeId string, aspectName string, methodName string, sign *types.Signature) Code {
	reflectMethodId := utils.Private(reflectTypeId + "Method" + methodName)
	return Blocks(
		Assign(
			Var(Ids(reflectMethodId, "_"), nil),
			Join(Id(reflectTypeId), Raw(".MethodByName("), Val(methodName), Raw(")")),
		),
		Method(
			Receiver("a", Id(aspectName)),
			methodName,
			generateAspectMethodSignature(sign),
			generateAspectMethodBody(reflectTypeId, reflectMethodId, methodName, sign),
		),
	)
}

func generateAspectMethodSignature(sign *types.Signature) Code {
	params := sign.Params()
	paramsNames := utils.Params(params.Len())
	n := params.Len()
	in := make([]Code, 0, n)
	if sign.Variadic() {
		n--
	}
	for i := 0; i < n; i++ {
		param := params.At(i)
		in = append(in, Param(paramsNames[i], tgen.Type(param.Type()), false))
	}
	if sign.Variadic() {
		param := params.At(n)
		in = append(in, Param(paramsNames[n], tgen.Type(param.Type()), true))
	}

	results := sign.Results()
	n = results.Len()
	out := make([]Code, 0, n)
	for i := 0; i < n; i++ {
		result := results.At(i)
		out = append(out, Param("", tgen.Type(result.Type()), false))
	}
	return Sign(In(in...), Out(out...))
}

func generateAspectMethodBody(
	reflectTypeId string,
	reflectMethodId string,
	name string,
	sign *types.Signature,
) Code {
	params := sign.Params()
	paramsNames := utils.Params(params.Len())
	results := sign.Results()
	resultsNames := utils.Results(results.Len())

	aspect := AssignAndDecl(Id("asp"), Call(Id("a.Factory.Aspect"), Ids(reflectTypeId, reflectMethodId)))

	paramsValues := make([]Code, 0, params.Len())
	for i, paramName := range paramsNames {
		param := params.At(i)
		paramsValues = append(paramsValues, Inst(
			SmartQual("aspect", "github.com/CherkashinEvgeny/goaspect", "Param"),
			Fields(
				Field("Name", Val(param.Name())),
				Field("Value", Id(paramName)),
			),
		))
	}
	before := Call(
		Raw("asp.Before"),
		Vals(paramsValues...),
	)

	var implCall Code
	if results.Len() == 0 {
		implCall = Call(
			Join(Raw("a.Impl."), Id(name)),
			Ids(paramsNames...),
		)
	} else {
		implCall = AssignAndDecl(
			Ids(resultsNames...),
			Call(
				Join(Raw("a.Impl."), Id(name)),
				Ids(paramsNames...),
			),
		)
	}

	resultValues := make([]Code, 0, results.Len())
	for i, resultName := range resultsNames {
		result := results.At(i)
		resultValues = append(resultValues, Inst(
			SmartQual("aspect", "github.com/CherkashinEvgeny/goaspect", "Param"),
			Fields(
				Field("Name", Val(result.Name())),
				Field("Value", Id(resultName)),
			),
		))
	}
	after := Call(
		Raw("asp.After"),
		Vals(resultValues...),
	)

	if results.Len() == 0 {
		return Blocks(
			aspect,
			before,
			implCall,
			after,
		)
	}
	returnResult := Return(Ids(resultsNames...))
	return Lines(
		aspect,
		before,
		implCall,
		after,
		returnResult,
	)
}
