// Copyright 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package swupd

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/clearlinux/mixer-tools/helpers"
)

// setFlagFromPathname sets the ModifierFlag for a file if it has a prefix in the input dirs file.
// If the dirs file does not exist, it is created with set of default values.
func (f *File) setFlagFromPathname(flag ModifierFlag) error {
	fileName := modifierFlagInfo[flag].HeuristicFile
	dirPaths, err := helpers.ReadFileAndSplit(modifierFlagInfo[flag].HeuristicFile)
	if os.IsNotExist(err) {

		heuristicFile, err := os.Create(fileName)
		if err != nil {
			return errors.Wrap(err, "Failed to create "+fileName)
		}

		dirPaths = modifierFlagInfo[flag].DefaultHeuristicDirs
		_, err = fmt.Fprintln(heuristicFile, dirPaths)
		if err != nil {
			return errors.Wrap(err, "Failed to write "+fileName)
		}

		err = heuristicFile.Close()
		if err != nil {
			return errors.Wrap(err, "Failed to close "+fileName)
		}

	} else if err != nil {
		return errors.Wrap(err, "Failed to read "+fileName)
	}

	for _, dirPath := range dirPaths {
		path := strings.TrimSpace(dirPath)
		if path == "" {
			continue
		}

		if strings.HasPrefix(f.Name, path) {
			f.Modifier = flag
			break
		}
	}

	return nil
}

func (m *Manifest) applyHeuristics() error {
	for _, f := range m.Files {
		if err := f.setModifierFromPathname(); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) setModifierFromPathname() error {
	// The order matters, first check for config, then state and then boot.
	// More important modifiers must happen last to overwrite earlier ones.
	var err error
	if err = f.setFlagFromPathname(ModifierConfig); err != nil {
		return err
	}
	if err := f.setFlagFromPathname(ModifierState); err != nil {
		return err
	}
	if err := f.setFlagFromPathname(ModifierBoot); err != nil {
		return err
	}
	if f.Status == StatusDeleted {
		f.Status = StatusGhosted
	}
	return nil
}
