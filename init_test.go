// Copyright 2019 Nick White.
// Use of this source code is governed by the GPLv3
// license that can be found in the LICENSE file.

package preproc

import (
	"flag"
	"os"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

// TestMain is needed to ensure flags are parsed
func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}
