package swagger

func RunExplainResource(args *ExplainResourceArgs) {
	spec := HackMustReadSwaggerSpecFromGithub(MustVersion(args.Version))
	spec.ResolveStructure(args)
}
