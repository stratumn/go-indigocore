// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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

package main

import (
	"context"
	"flag"

	"github.com/stratumn/go-core/bcbatchfossilizer"
	"github.com/stratumn/go-core/blockchain/dummytimestamper"
	"github.com/stratumn/go-core/fossilizer/fossilizerhttp"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/util"
)

var (
	version = "x.x.x"
	commit  = "00000000000000000000000000000000"
)

func init() {
	fossilizerhttp.RegisterFlags()
	bcbatchfossilizer.RegisterFlags()
	monitoring.RegisterFlags()
}

func main() {
	flag.Parse()

	ctx := context.Background()
	ctx = util.CancelOnInterrupt(ctx)

	a := monitoring.NewFossilizerAdapter(
		bcbatchfossilizer.RunWithFlags(ctx, version, commit, dummytimestamper.Timestamper{}),
		"dummybatchfossilizer",
	)
	fossilizerhttp.RunWithFlags(ctx, a)
}
