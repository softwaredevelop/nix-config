//revive:disable:package-comments,exported
package main

import (
	"github.com/pulumi/pulumi-github/sdk/v6/go/github"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

const (
	repositoryName        = "nix-config"
	repositoryDescription = "Declarative NixOS configuration for reproducible system builds and development environments."
	defaultBranch         = "main"
)

// GitHubResources holds the created GitHub resources, making them accessible for testing and exporting.
type GitHubResource struct {
	Repository *github.Repository
}

// defineInfrastructure defines the GitHub resources for the project.
// It is separated from main() to be independently testable.
func defineInfrastructure(ctx *pulumi.Context) (*GitHubResource, error) {
	repository, err := github.NewRepository(ctx, "nixConfigRepository", &github.RepositoryArgs{
		DeleteBranchOnMerge: pulumi.Bool(true),
		Description:         pulumi.String(repositoryDescription),
		HasIssues:           pulumi.Bool(true),
		HasProjects:         pulumi.Bool(true),
		Name:                pulumi.String(repositoryName),
		Topics: pulumi.StringArray{
			pulumi.String("dagger"),
			pulumi.String("dotfiles"),
			pulumi.String("github"),
			pulumi.String("gitlab"),
			pulumi.String("golang"),
			pulumi.String("nix-dotfiles"),
			pulumi.String("nix"),
			pulumi.String("nixos"),
			pulumi.String("pulumi"),
			pulumi.String("vscode"),
		},
		Visibility: pulumi.String("public"),
		// VulnerabilityAlerts: pulumi.Bool(true),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	_, err = github.NewBranchProtection(ctx, "nixConfigMainBranchProtection", &github.BranchProtectionArgs{
		RepositoryId:          repository.NodeId,
		Pattern:               pulumi.String(defaultBranch),
		RequiredLinearHistory: pulumi.Bool(true),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	_, err = github.NewIssueLabel(ctx, "nixConfigLabelGithubActions", &github.IssueLabelArgs{
		Color:       pulumi.String("E66E01"),
		Description: pulumi.String("This issue is related to github-actions dependencies"),
		Name:        pulumi.String("dependencies:github-actions"),
		Repository:  repository.Name,
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	_, err = github.NewIssueLabel(ctx, "nixConfigLabelGoModules", &github.IssueLabelArgs{
		Color:       pulumi.String("9BE688"),
		Description: pulumi.String("This issue is related to go modules dependencies"),
		Name:        pulumi.String("dependencies:go-modules"),
		Repository:  repository.Name,
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	gitlabRepo := config.Require(ctx, "gitlabRepository")
	gitlabToken := config.RequireSecret(ctx, "gitlabToken")
	gitlabOwner := config.Require(ctx, "gitlabOwner")

	_, err = github.NewActionsSecret(ctx, "nixConfigGitlabRepositorySecret", &github.ActionsSecretArgs{
		Repository:     repository.Name,
		SecretName:     pulumi.String("GITLAB_REPOSITORY"),
		PlaintextValue: pulumi.String(gitlabRepo),
	}, pulumi.Parent(repository), pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	_, err = github.NewActionsSecret(ctx, "nixConfigGitlabTokenSecret", &github.ActionsSecretArgs{
		Repository:     repository.Name,
		SecretName:     pulumi.String("GITLAB_TOKEN"),
		PlaintextValue: gitlabToken,
	}, pulumi.Parent(repository), pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	_, err = github.NewActionsSecret(ctx, "nixConfigGitlabOwnerSecret", &github.ActionsSecretArgs{
		Repository:     repository.Name,
		SecretName:     pulumi.String("GITLAB_OWNER"),
		PlaintextValue: pulumi.String(gitlabOwner),
	}, pulumi.Parent(repository), pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	return &GitHubResource{
		Repository: repository,
	}, nil
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		resources, err := defineInfrastructure(ctx)
		if err != nil {
			return err
		}

		// Export outputs from the returned resources
		ctx.Export("repository", resources.Repository.Name)
		ctx.Export("repositoryUrl", resources.Repository.HtmlUrl)
		return nil
	})
}
