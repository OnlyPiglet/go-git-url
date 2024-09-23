package v1

import (
	"fmt"
	giturl "github.com/chainguard-dev/git-urls"
	"github.com/kubescape/go-git-url/apis"
	"github.com/kubescape/go-git-url/apis/gitlabapi"
	"net/url"
	"os"
	"strings"
)

type GiteeURL struct {
	host    string
	owner   string // repo owner
	repo    string // repo name
	project string
	branch  string
	path    string
	token   string // github token
	isFile  bool

	gitLabAPI gitlabapi.IGitLabAPI
}

// NewGitHubParserWithURL parsed instance of a github parser
func NewGitLabParserWithURL(host, fullURL string) (*GiteeURL, error) {
	gl := &GiteeURL{
		gitLabAPI: gitlabapi.NewGitLabAPI(host),
		token:     os.Getenv("GITLAB_TOKEN"),
	}

	if err := gl.Parse(fullURL); err != nil {
		return nil, err
	}

	return gl, nil
}

func (gl *GiteeURL) GetURL() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   gl.GetHostName(),
		Path:   fmt.Sprintf("%s/%s", gl.GetOwnerName(), gl.GetRepoName()),
	}
}

func IsHostGitee(host string) bool { return strings.Contains(host, "gitee") }

func (gl *GiteeURL) GetProvider() string    { return apis.ProviderGitLab.String() }
func (gl *GiteeURL) GetHostName() string    { return gl.host }
func (gl *GiteeURL) GetProjectName() string { return gl.project }
func (gl *GiteeURL) GetBranchName() string  { return gl.branch }
func (gl *GiteeURL) GetOwnerName() string   { return gl.owner }
func (gl *GiteeURL) GetRepoName() string    { return gl.repo }
func (gl *GiteeURL) GetPath() string        { return gl.path }
func (gl *GiteeURL) GetToken() string       { return gl.token }
func (gl *GiteeURL) GetHttpCloneURL() string {
	return fmt.Sprintf("https://%s/%s/%s.git", gl.host, gl.owner, gl.repo)
}

func (gl *GiteeURL) SetOwnerName(o string)         { gl.owner = o }
func (gl *GiteeURL) SetProjectName(project string) { gl.project = project }
func (gl *GiteeURL) SetRepoName(r string)          { gl.repo = r }
func (gl *GiteeURL) SetBranchName(branch string)   { gl.branch = branch }
func (gl *GiteeURL) SetPath(p string)              { gl.path = p }
func (gl *GiteeURL) SetToken(token string)         { gl.token = token }

// Parse URL
func (gl *GiteeURL) Parse(fullURL string) error {
	parsedURL, err := giturl.Parse(fullURL)
	if err != nil {
		return err
	}

	gl.host = parsedURL.Host

	index := 0

	splittedRepo := strings.FieldsFunc(parsedURL.Path, func(c rune) bool { return c == '/' }) // trim empty fields from returned slice
	if len(splittedRepo) < 2 {
		return fmt.Errorf("expecting <user>/<repo> in url path, received: '%s'", parsedURL.Path)
	}

	// in gitlab <user>/<repo> are separated from blob/tree/raw with a -
	for i := range splittedRepo {
		if splittedRepo[i] == "-" {
			break
		}
		index = i
	}

	gl.owner = strings.Join(splittedRepo[:index], "/")
	gl.repo = strings.TrimSuffix(splittedRepo[index], ".git")
	index++

	// root of repo
	if len(splittedRepo) < index+1 {
		return nil
	}

	if splittedRepo[index] == "-" {
		index++ // skip "-" symbol in URL
	}

	// is file or dir
	switch splittedRepo[index] {
	case "blob", "tree", "raw":
		index++
	}

	if len(splittedRepo) < index+1 {
		return nil
	}

	gl.branch = splittedRepo[index]
	index += 1

	if len(splittedRepo) < index+1 {
		return nil
	}
	gl.path = strings.Join(splittedRepo[index:], "/")

	return nil
}

// Set the default brach of the repo
func (gl *GiteeURL) SetDefaultBranchName() error {
	branch, err := gl.gitLabAPI.GetDefaultBranchName(gl.GetOwnerName(), gl.GetRepoName(), gl.headers())
	if err != nil {
		return err
	}
	gl.branch = branch
	return nil
}

func (gl *GiteeURL) headers() *gitlabapi.Headers {
	return &gitlabapi.Headers{Token: gl.GetToken()}
}
