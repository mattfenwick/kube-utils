package swagger

import (
	"fmt"
	"github.com/mattfenwick/kube-utils/go/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sort"
	"strings"
)

func RunExplain(args *ExplainArgs) {
	swaggerSpec := MustReadSwaggerSpec(MustVersion(args.Version))

	// no types specified?  use them all
	//   otherwise, filter down to just the ones requested
	var typeNames []string
	if len(args.TypeNames) == 0 {
		for name := range swaggerSpec.DefinitionsByNameByGroup() {
			typeNames = append(typeNames, name)
		}
		sort.Strings(typeNames)
	} else {
		typeNames = args.TypeNames // TODO should this be sorted, or respect the input order?
	}

	for _, typeName := range typeNames {
		logrus.Debugf("analysing type %s", typeName)
		analyses := swaggerSpec.AnalyzeType(typeName)

		// no group/versions specified?  use them all
		//   otherwise, filter down to just the ones requested
		if len(args.GroupVersions) > 0 {
			filteredAnalyses := map[string]interface{}{}
			for _, groupVersion := range args.GroupVersions {
				if analysis, ok := analyses[groupVersion]; ok {
					filteredAnalyses[groupVersion] = analysis
				} else {
					logrus.Debugf("type %s not found under group/version %s (%+v)", typeName, groupVersion, utils.SortedKeys(analyses))
				}
			}
			analyses = filteredAnalyses
		}

		gvks := utils.SortedKeys(analyses)
		if len(gvks) == 0 {
			logrus.Debugf("no group/versions found for %s", typeName)
			continue
		}
		for _, groupVersion := range gvks {
			analysis := analyses[groupVersion]
			switch args.Format {
			case "table":
				fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, AnalysisTypeTable(analysis))
			case "condensed":
				fmt.Printf("%s.%s:\n%s\n", groupVersion, typeName, strings.Join(AnalysisTypeSummary(analysis), "\n"))
			default:
				panic(errors.Errorf("invalid output format: %s", args.Format))
			}
			fmt.Println()
		}
		fmt.Println()
	}
}
