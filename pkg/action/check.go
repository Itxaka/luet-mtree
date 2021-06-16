/*
Copyright © 2021 SUSE LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package action

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/itxaka/luet-mtree/pkg/log"
	"github.com/vbatts/go-mtree"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type checkAction struct {
	target         string
	validationFile string
	format         string
	exclude        []string
}

func NewCheckAction(t string, v string, f string, x []string) *checkAction {
	return &checkAction{target: t, validationFile: v, format: f, exclude: x}
}

func (action checkAction) Run() (string, error) {
	log.Log("Checking %s against validation file %s", action.target, action.validationFile)
	if len(action.exclude) > 0 {
		log.Log("Using the following exclude list: %v", action.exclude)
	}
	spec := &mtree.DirectoryHierarchy{}
	stateDh := &mtree.DirectoryHierarchy{}
	var err error
	var excludes []mtree.ExcludeFunc
	var res []mtree.InodeDelta
	var cleanRes []mtree.InodeDelta

	fh, err := os.Open(action.validationFile)
	if err != nil {
		return "", err
	}
	spec, err = mtree.ParseSpec(fh)
	err = fh.Close()
	if err != nil {
		return "", err
	}

	stateKeyworks := spec.UsedKeywords()

	// If its not a dir, try to uncompress
	info, _ := os.Stat(action.target)
	if !info.IsDir() {
		uncompressedTar, err := unCompress(action.target)
		ts := mtree.NewTarStreamer(uncompressedTar, excludes, stateKeyworks)
		if _, err := io.Copy(ioutil.Discard, ts); err != nil && err != io.EOF {
			return "", err
		}
		if err := ts.Close(); err != nil {
			return "", err
		}
		stateDh, err = ts.Hierarchy()
		if err != nil {
			return "", err
		}

	} else {
		stateDh, err = mtree.Walk(action.target, excludes, stateKeyworks, nil)
		if err != nil {
			return "", err
		}
	}

	res, err = mtree.Compare(spec, stateDh, stateKeyworks)
	if err != nil {
		return "", err
	}

	// Skip excluded paths from results
	// This is very useful for cache directories (i.e. luet cache!) or post-install dirs which were not part of the
	// build process (i.e. OEM configs, cloud-init stuff)
	for _, diff := range res {
		// No excludes, return the full list
		if len(action.exclude) == 0 {
			cleanRes = res
		} else {
			// Got excludes, lets check them!
			if findInSlice(action.exclude, diff.Path()) {
				// Oh my! we matched! Log and skip the diff for that path
				log.Log("Path %s found against exclude values %s, skipping entry.", diff.Path(), action.exclude)
			} else {
				// We didnt match the excludes, add it to the results
				cleanRes = append(cleanRes, diff)
			}
		}
	}

	out := formats[action.format](cleanRes)

	for _, diff := range cleanRes {
		if diff.Type() == mtree.Modified || diff.Type() == mtree.Missing || diff.Type() == mtree.Extra {
			return out, errors.New("validation failed")
		}
	}
	return out, nil
}

func findInSlice(slice []string, val string) bool {
	for _, item := range slice {
		if strings.HasPrefix(val, item) {
			return true
		}
	}
	return false
}

var formats = map[string]func([]mtree.InodeDelta) string{
	// Outputs the errors in the BSD format.
	"bsd": func(d []mtree.InodeDelta) string {
		var buffer bytes.Buffer
		for _, delta := range d {
			_, _ = fmt.Fprintln(&buffer, delta)
		}
		return buffer.String()
	},

	// Outputs the full result struct in JSON.
	"json": func(d []mtree.InodeDelta) string {
		var buffer bytes.Buffer
		if err := json.NewEncoder(&buffer).Encode(d); err != nil {
			panic(err)
		}
		return buffer.String()
	},

	// Outputs only the paths which failed to validate.
	"path": func(d []mtree.InodeDelta) string {
		var buffer bytes.Buffer
		for _, delta := range d {
			if delta.Type() == mtree.Modified {
				_, _ = fmt.Fprintln(&buffer, delta.Path())
			}
		}
		return buffer.String()
	},
}
