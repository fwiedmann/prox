package configure

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/fwiedmann/prox/domain/entity/route"
	"gopkg.in/yaml.v2"
)

var (
	ErrorInvalidFileType = errors.New("given file type is invalid, only .yaml or yml is allowed")
)

func NewFileConfigureUseCase(f string, manager route.Configurator) UseCase {
	return &file{
		pathToFile:   f,
		routeManager: manager,
	}
}

type file struct {
	pathToFile   string
	routeManager route.Configurator
}

func (f *file) StartConfigure(ctx context.Context, errChan chan<- error) {
	file, err := os.Open(f.pathToFile)
	if err != nil {
		errChan <- err
		return
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		errChan <- err
		return
	}

	if !strings.HasSuffix(file.Name(), ".yaml") && !strings.HasSuffix(file.Name(), ".yaml") {
		errChan <- ErrorInvalidFileType
	}

	routes := make([]*route.Route, 0)
	if err := yaml.Unmarshal(content, &routes); err != nil {
		errChan <- err
		return
	}

	log.Debugf("Parsed routes config file \"%s\": %v", f.pathToFile, routes)

	if !hasDuplicates(routes) {
		for _, r := range routes {
			if err := f.routeManager.CreateRoute(ctx, r); err != nil {
				if !errors.Is(err, route.ErrorAlreadyExists) {
					log.Errorf("could not create route with name %s, error: %s", r.NameID, err)
					continue
				}
				if err := f.routeManager.UpdateRoute(ctx, r); err != nil {
					log.Errorf("could not update route with name %s, error: %s", r.NameID, err)
					continue
				}
			}
		}
		log.Info("Successfully configured proxy")
	}

	initStat, err := os.Stat(f.pathToFile)
	if err != nil {
		errChan <- err
	}

	for {
		if ctx.Err() != nil {
			return
		}

		stat, err := os.Stat(f.pathToFile)
		if err != nil {
			errChan <- err
		}

		if initStat.ModTime() != stat.ModTime() {
			log.Info("Routes configuration file update noticed, will reload")
			f.StartConfigure(ctx, errChan)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func hasDuplicates(routes []*route.Route) bool {
	var hasDuplicates bool
	duplicates := make(map[string]int)

	for _, r1 := range routes {
		count := 0
		for _, r2 := range routes {
			if r1.NameID == r2.NameID {
				count++
			}
		}
		if count > 1 {
			hasDuplicates = true
			duplicates[string(r1.NameID)] = count
		}
	}
	if hasDuplicates {
		log.Error("configuration has duplicated route names:")
		for key, val := range duplicates {
			log.Errorf("route-name: \"%s\", count: %d", key, val)
		}
	}
	return hasDuplicates
}
