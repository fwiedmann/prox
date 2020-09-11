package config

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/ghodss/yaml"

	log "github.com/sirupsen/logrus"
)

// TLS config
type TLS struct {
	configFile string
	certStore  map[string]tls.Certificate
	mtx        sync.RWMutex
}

// NewDynamicTLSConfig
func NewDynamicTLSConfig(configFile string) *TLS {
	return &TLS{
		configFile: configFile,
		certStore:  make(map[string]tls.Certificate),
	}
}

// Pair hold the paths to a certificate pair
type Pair struct {
	Certificate string `yaml="certificate"`
	Key         string `yaml="key"`
}

func (p Pair) ID() string {
	return fmt.Sprintf("ID_%s_%s", p.Certificate, p.Key)
}

// GetCertificate
func (t *TLS) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	for _, cert := range t.certStore {
		if err := hello.SupportsCertificate(&cert); err == nil {
			return &cert, nil
		}
	}

	return nil, errors.New("not found")
}

func (t *TLS) StartWatch(ctx context.Context, errChan chan<- error) {
	file, err := ioutil.ReadFile(t.configFile)
	if err != nil {
		errChan <- err
	}

	pairs := make([]Pair, 0)
	if err := yaml.Unmarshal(file, &pairs); err != nil {
		errChan <- err
	}
	log.Debugf("Parsed tls config file \"%s\": %+v", t.configFile, pairs)

	t.deleteStalePairs(pairs)
	t.startPairWatchers(ctx, pairs)

	initConfigFileStat, err := os.Stat(t.configFile)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Successfully configured tls configuration")

	for {
		configFileStat, err := os.Stat(t.configFile)
		if err != nil {
			log.Error(err)
			return
		}

		if initConfigFileStat.ModTime() != configFileStat.ModTime() {
			log.Info("Routes configuration file update noticed, will reload")
			t.StartWatch(ctx, errChan)
		}

		if ctx.Err() != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

}

func (t *TLS) startWatchPair(ctx context.Context, pair Pair) {
	cer, err := tls.LoadX509KeyPair(pair.Certificate, pair.Key)
	if err != nil {
		log.Error(err)
		return
	}

	t.mtx.Lock()
	t.certStore[pair.ID()] = cer
	t.mtx.Unlock()

	initCertStat, err := os.Stat(pair.Certificate)
	if err != nil {
		log.Error(err)
		return
	}
	initKeyStat, err := os.Stat(pair.Key)
	if err != nil {
		log.Error(err)
		return
	}

	for {
		certStat, err := os.Stat(pair.Certificate)
		if err != nil {
			log.Error(err)
			if errors.Is(err, os.ErrNotExist) {
				t.mtx.Lock()
				delete(t.certStore, pair.ID())
				t.mtx.Unlock()
				return
			}
		}
		keyStat, err := os.Stat(pair.Key)
		if err != nil {
			log.Error(err)
			if errors.Is(err, os.ErrNotExist) {
				t.mtx.Lock()
				delete(t.certStore, pair.ID())
				t.mtx.Unlock()
				return
			}
		}

		if initCertStat.ModTime() != certStat.ModTime() {
			t.startWatchPair(ctx, pair)
		}

		if initKeyStat.ModTime() != keyStat.ModTime() {
			t.startWatchPair(ctx, pair)
		}

		t.mtx.RLock()
		if _, ok := t.certStore[pair.ID()]; !ok {
			t.mtx.RUnlock()
			return
		}
		t.mtx.RUnlock()

		if ctx.Err() != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (t *TLS) deleteStalePairs(pairs []Pair) {
	t.mtx.Lock()
	for id, _ := range t.certStore {
		var found bool
		for _, pair := range pairs {
			if id == pair.ID() {
				found = true
			}
		}
		if !found {
			delete(t.certStore, id)
		}
	}
	t.mtx.Unlock()
}

func (t *TLS) startPairWatchers(ctx context.Context, pairs []Pair) {
	for _, pair := range pairs {
		if _, ok := t.certStore[pair.ID()]; !ok {
			go t.startWatchPair(ctx, pair)
		}
	}
}
