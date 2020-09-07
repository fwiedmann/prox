package route

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"
)

var (
	ErrorEmptyRoute                      = errors.New("route is nil")
	ErrorNoEntityID                      = errors.New("route NameID is empty")
	ErrorEmptyRequestIdentifiers         = errors.New("all routes RequestIdentifier are empty. At least one is required")
	ErrorDuplicatedRequestIdentifier     = errors.New("all request identifiers are configured. only one per route host / path is allowed")
	ErrorDuplicatedHostRequestIdentifier = errors.New("all host request identifiers are configured. only one per route is allowed")
	ErrorDuplicatedPathRequestIdentifier = errors.New("all path request identifiers are configured. only one per route is allowed")
	ErrorInvalidHostName                 = errors.New(fmt.Sprintf("hostname is invalid. Used expression: %s", hostNameRegexp.String()))
	ErrorInvalidCacheTimeOutDuration     = errors.New("invalid cache time out duration format")
	ErrorInvalidUpstreamTimeOutDuration  = errors.New("invalid upstream time out duration format")

	hostNameRegexp = regexp.MustCompile("^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])(\\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9]))*$")
	pathRegexp     = regexp.MustCompile("(\\/[0-9].*\\?|$)")
	wildcardRegexp = regexp.MustCompile("[\\s\\S]*")
)

const defaultHTTPUpstreamTimeoutDuration = "10s"
const defaultCacheTimeoutDuration = "10m"
const megaBytesToBytesMultiplier = 1e+6

type manager struct {
	repo repository
}

// NewManager return a manager to interact with the entities stored in the repository.
// The manager is responsible to apply the Route business rules.
func NewManager(r repository) Manager {
	return &manager{
		repo: r,
	}
}

// UpdateRoute which is stored in the managers repository. If the context has an error UpdateRoute will not call the repository and will return.
func (m *manager) UpdateRoute(ctx context.Context, r *Route) error {
	if r == nil {
		return ErrorEmptyRoute
	}

	if r.NameID == "" {
		return ErrorNoEntityID
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return m.repo.UpdateRoute(ctx, r)
}

// ListRoutes which are stored in the managers repository. If the context has an error UpdateRoute will not call the repository and will return.
func (m *manager) ListRoutes(ctx context.Context) ([]*Route, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return m.repo.ListRoutes(ctx)
}

// CreateRoute in the managers repository. If the context has an error CreateRoute will not call the repository and will return.
// CreateRoute also responsible to validate the given Route.
func (m *manager) CreateRoute(ctx context.Context, r *Route) error {
	if r == nil {
		return ErrorEmptyRoute
	}

	if r.NameID == "" {
		return ErrorNoEntityID
	}

	if err := parseDurations(r); err != nil {
		return err
	}

	parseCacheMaxBodySize(r)

	if err := validateRouteRequestIdentifiers(r); err != nil {
		return err
	}

	if err := configureRouteRequestMatches(r); err != nil {
		return err
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}
	return m.repo.CreateRoute(ctx, r)
}

func parseDurations(r *Route) error {
	if r.CacheTimeOutDuration == "" {
		r.CacheTimeOutDuration = defaultCacheTimeoutDuration
	}

	cacheTimeOut, err := time.ParseDuration(r.CacheTimeOutDuration)
	if err != nil {
		return ErrorInvalidCacheTimeOutDuration
	}
	r.cacheTimeOutDuration = cacheTimeOut

	if r.UpstreamTimeoutDuration == "" {
		r.UpstreamTimeoutDuration = defaultHTTPUpstreamTimeoutDuration
	}

	upstreamTimeOut, err := time.ParseDuration(r.UpstreamTimeoutDuration)
	if err != nil {
		return ErrorInvalidUpstreamTimeOutDuration
	}
	r.upstreamTimeoutDuration = upstreamTimeOut
	return nil
}

func parseCacheMaxBodySize(r *Route) {
	if r.CacheMaxBodySizeInMegaBytes <= 0 {
		r.CacheMaxBodySizeInMegaBytes = -1
	}

	r.cacheMaxBodySizeInBytes = r.CacheMaxBodySizeInMegaBytes * megaBytesToBytesMultiplier
}

func validateRouteRequestIdentifiers(r *Route) error {
	if r.Hostname == "" && r.HostnameRegexp == "" && r.Path == "" && r.PathRegexp == "" {
		return ErrorEmptyRequestIdentifiers
	}

	if r.Hostname != "" && r.HostnameRegexp != "" && r.Path != "" && r.PathRegexp != "" {
		return ErrorDuplicatedRequestIdentifier
	}

	if r.Hostname != "" && r.HostnameRegexp != "" {
		return ErrorDuplicatedHostRequestIdentifier
	}

	if r.Path != "" && r.PathRegexp != "" {
		return ErrorDuplicatedPathRequestIdentifier
	}
	return nil
}

func configureRouteRequestMatches(r *Route) error {
	hostMatchRegexp, err := getRouteHostMatch(string(r.Hostname), string(r.HostnameRegexp))
	if err != nil {
		return err
	}
	r.hostMatch = hostMatchRegexp

	pathMatchRegexp, err := getRoutePathMatch(string(r.Path), string(r.PathRegexp))
	if err != nil {
		return err
	}
	r.pathMatch = pathMatchRegexp

	return nil
}

func getRouteHostMatch(host, hostExpr string) (*regexp.Regexp, error) {
	if host != "" && hostExpr == "" {
		expr, err := addHostRegexpStartAndEndPosition(host)
		if err != nil {
			return nil, err
		}

		regex, err := regexp.Compile(expr)
		if err != nil {
			return nil, err
		}
		return regex, nil
	}

	if host == "" && hostExpr != "" {
		regex, err := regexp.Compile(hostExpr)
		if err != nil {
			return nil, err
		}
		return regex, nil
	}

	return wildcardRegexp, nil
}

func addHostRegexpStartAndEndPosition(s string) (string, error) {
	if !hostNameRegexp.MatchString(s) {
		return "", ErrorInvalidHostName
	}
	return fmt.Sprintf("^%s$", s), nil
}

func getRoutePathMatch(path, pathExpr string) (*regexp.Regexp, error) {
	if path != "" && pathExpr == "" {
		regex, err := regexp.Compile(path)
		if err != nil {
			return nil, err
		}
		return regex, nil
	}

	if path == "" && pathExpr != "" {
		regex, err := regexp.Compile(pathExpr)
		if err != nil {
			return nil, err
		}
		return regex, nil
	}
	return wildcardRegexp, nil
}

// DeleteRoute which is stored in the managers repository. If the context has an error DeleteRoute will not call the repository and will return.
func (m *manager) DeleteRoute(ctx context.Context, id NameID) error {
	if id == "" {
		return ErrorNoEntityID
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}
	return m.repo.DeleteRoute(ctx, id)
}
