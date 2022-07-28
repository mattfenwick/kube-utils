package swagger

func HackExplainResource(args *ExplainResourceArgs) {
	spec := HackMustReadSwaggerSpecFromGithub(MustVersion(args.Version))
	spec.ResolveStructure(args)
}
