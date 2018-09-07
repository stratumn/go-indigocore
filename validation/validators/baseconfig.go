// Copyright 2017 Stratumn SAS. All rights reserved.
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

package validators

import (
	"github.com/pkg/errors"
	"github.com/stratumn/go-chainscript"
)

var (
	// ErrMissingProcess is returned when the process name is missing for validation.
	ErrMissingProcess = errors.New("validator requires a process")

	// ErrMissingLinkStep is returned when the link step is missing for validation.
	ErrMissingLinkStep = errors.New("validator requires a link step")
)

// ValidatorBaseConfig is used to identify a validator by its process and link
// step.
// Its ShouldValidate method is called by all validators.
type ValidatorBaseConfig struct {
	Process  string
	LinkStep string
}

// NewValidatorBaseConfig returns a new ValidatorBaseConfig.
func NewValidatorBaseConfig(process, linkStep string) (*ValidatorBaseConfig, error) {
	if len(process) == 0 {
		return nil, ErrMissingProcess
	}

	if len(linkStep) == 0 {
		return nil, ErrMissingLinkStep
	}
	return &ValidatorBaseConfig{Process: process, LinkStep: linkStep}, nil
}

// ShouldValidate returns true if the link matches the validator's process
// and type. Otherwise the link is considered valid because this validator
// doesn't apply to it.
func (bv *ValidatorBaseConfig) ShouldValidate(link *chainscript.Link) bool {
	if link.Meta.Process.Name != bv.Process {
		return false
	}

	if link.Meta.Step != bv.LinkStep {
		return false
	}

	return true
}
