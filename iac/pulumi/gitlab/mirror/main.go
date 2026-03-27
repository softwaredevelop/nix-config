//revive:disable:package-comments,exported
package main

import (
	"fmt"

	"github.com/pulumi/pulumi-gitlab/sdk/v6/go/gitlab"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	gitlabGroupPath = "mirror-e/github-softwaredevelop"
	projectName     = "nix-config"
)

type GitLabResource struct {
	Project *gitlab.Project
}

func defineInfrastructure(ctx *pulumi.Context) (*GitLabResource, error) {
	groupID, err := gitlab.LookupGroup(ctx, &gitlab.LookupGroupArgs{
		FullPath: new(gitlabGroupPath),
	}, nil)
	if err != nil {
		return nil, err
	}

	projectDescription := fmt.Sprintf("A GitLab project for mirroring %s GitHub repository.", projectName)
	project, err := gitlab.NewProject(ctx, "nixConfigProject", &gitlab.ProjectArgs{
		AutoCancelPendingPipelines:       pulumi.String("enabled"),
		BuildsAccessLevel:                pulumi.String("private"),
		Description:                      pulumi.String(projectDescription),
		IssuesEnabled:                    pulumi.Bool(true),
		LfsEnabled:                       pulumi.Bool(true),
		MergeMethod:                      pulumi.String("merge"),
		MergeRequestsEnabled:             pulumi.Bool(true),
		Name:                             pulumi.String(projectName),
		NamespaceId:                      pulumi.Int(groupID.GroupId),
		OnlyAllowMergeIfPipelineSucceeds: pulumi.Bool(true),
		RemoveSourceBranchAfterMerge:     pulumi.Bool(true),
		SharedRunnersEnabled:             pulumi.Bool(true),
		Topics: pulumi.StringArray{
			pulumi.String("dagger"),
			pulumi.String("dotfiles"),
			pulumi.String("github"),
			pulumi.String("gitlab"),
			pulumi.String("golang"),
			pulumi.String("mirror"),
			pulumi.String("nix-dotfiles"),
			pulumi.String("nix"),
			pulumi.String("nixos"),
			pulumi.String("pulumi"),
			pulumi.String("vscode"),
		},
		VisibilityLevel: pulumi.String("private"),
		// VulnerabilityAlerts: pulumi.Bool(true),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	return &GitLabResource{
		Project: project,
	}, nil
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		resources, err := defineInfrastructure(ctx)
		if err != nil {
			return err
		}

		ctx.Export("projectName", resources.Project.Name)
		ctx.Export("projectWebUrl", resources.Project.WebUrl)

		return nil
	})
}
