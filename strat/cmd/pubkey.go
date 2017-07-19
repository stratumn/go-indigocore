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

package cmd

// PGP public key used to verify updates.
const pubKey = `
-----BEGIN PGP PUBLIC KEY BLOCK-----

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
=kKDW
-----END PGP PUBLIC KEY BLOCK-----
`
