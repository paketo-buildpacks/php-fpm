package phpfpm

import "github.com/paketo-buildpacks/packit/v2"

func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Requires: []packit.BuildPlanRequirement{
					{
						Name: PhpDist,
						Metadata: map[string]interface{}{
							"build":  true,
							"launch": true,
						},
					},
				},
				Provides: []packit.BuildPlanProvision{
					{
						Name: PhpFpmDependency,
					},
				},
			},
		}, nil
	}
}
