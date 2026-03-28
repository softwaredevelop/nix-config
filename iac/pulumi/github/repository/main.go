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

// Config layer

type GitLabConfig struct {
	Repository string
	Owner      string
	Token      pulumi.StringOutput
}

func LoadGitLabConfig(ctx *pulumi.Context) GitLabConfig {
	cfg := config.New(ctx, "gitlab")

	return GitLabConfig{
		Repository: cfg.Require("repository"),
		Owner:      cfg.Require("owner"),
		Token:      cfg.RequireSecret("token"),
	}
}

// Resources

type GitHubResource struct {
	Repository *github.Repository
}

func defineInfrastructure(ctx *pulumi.Context) (*GitHubResource, error) {
	// Load config
	gitlab := LoadGitLabConfig(ctx)

	// Repository
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

	// Branch protection
	_, err = github.NewBranchProtection(ctx, "nixConfigMainBranchProtection", &github.BranchProtectionArgs{
		RepositoryId:          repository.NodeId,
		Pattern:               pulumi.String(defaultBranch),
		RequiredLinearHistory: pulumi.Bool(true),
	}, pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	// Labels
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

	// GitHub Actions secrets
	_, err = github.NewActionsSecret(ctx, "nixConfigGitlabRepositorySecret", &github.ActionsSecretArgs{
		Repository:     repository.Name,
		SecretName:     pulumi.String("GITLAB_REPOSITORY"),
		PlaintextValue: pulumi.String(gitlab.Repository),
	}, pulumi.Parent(repository), pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	_, err = github.NewActionsSecret(ctx, "nixConfigGitlabTokenSecret", &github.ActionsSecretArgs{
		Repository:     repository.Name,
		SecretName:     pulumi.String("GITLAB_TOKEN"),
		PlaintextValue: gitlab.Token,
	}, pulumi.Parent(repository), pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	_, err = github.NewActionsSecret(ctx, "nixConfigGitlabOwnerSecret", &github.ActionsSecretArgs{
		Repository:     repository.Name,
		SecretName:     pulumi.String("GITLAB_OWNER"),
		PlaintextValue: pulumi.String(gitlab.Owner),
	}, pulumi.Parent(repository), pulumi.Protect(false))
	if err != nil {
		return nil, err
	}

	return &GitHubResource{
		Repository: repository,
	}, nil
}

// Entry point

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		resources, err := defineInfrastructure(ctx)
		if err != nil {
			return err
		}

		ctx.Export("repositoryName", resources.Repository.Name)
		ctx.Export("repositoryUrl", resources.Repository.HtmlUrl)
		return nil
	})
}
