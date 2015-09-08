// Copyright 2015 Google Inc. All Rights Reserved.
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

package integration_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"runtime"
	"testing"

	"github.com/googlecloudplatform/gcsfuse/internal/canned"
	. "github.com/jacobsa/oglematchers"
	. "github.com/jacobsa/ogletest"
)

func TestMountHelper(t *testing.T) { RunTests(t) }

////////////////////////////////////////////////////////////////////////
// Boilerplate
////////////////////////////////////////////////////////////////////////

type MountHelperTest struct {
	// Path to the mount(8) helper binary.
	helperPath string

	// A temporary directory into which a file system may be mounted. Removed in
	// TearDown.
	dir string
}

var _ SetUpInterface = &MountHelperTest{}
var _ TearDownInterface = &MountHelperTest{}

func init() { RegisterTestSuite(&MountHelperTest{}) }

func (t *MountHelperTest) SetUp(_ *TestInfo) {
	var err error

	// Set up the appropriate helper path.
	switch runtime.GOOS {
	case "darwin":
		t.helperPath = path.Join(gBuildDir, "sbin/mount_gcsfuse")

	case "linux":
		t.helperPath = path.Join(gBuildDir, "sbin/mount.gcsfuse")

	default:
		AddFailure("Don't know how to deal with OS: %q", runtime.GOOS)
		AbortTest()
	}

	// Set up the temporary directory.
	t.dir, err = ioutil.TempDir("", "mount_helper_test")
	AssertEq(nil, err)
}

func (t *MountHelperTest) TearDown() {
	err := os.Remove(t.dir)
	AssertEq(nil, err)
}

func (t *MountHelperTest) mount(args []string) (err error) {
	cmd := exec.Command(t.helperPath)
	cmd.Args = append(cmd.Args, args...)
	cmd.Env = []string{
		fmt.Sprintf("PATH=%s", path.Join(gBuildDir, "bin")),
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("CombinedOutput: %v\nOutput:\n%s", err, output)
		return
	}

	return
}

////////////////////////////////////////////////////////////////////////
// Tests
////////////////////////////////////////////////////////////////////////

func (t *MountHelperTest) BadUsage() {
	testCases := []struct {
		args           []string
		expectedOutput string
	}{
		// Too few args
		0: {
			[]string{canned.FakeBucketName},
			"two positional arguments",
		},

		// Too many args
		1: {
			[]string{canned.FakeBucketName, "a", "b"},
			"Unexpected arg 3",
		},

		// Trailing -o
		2: {
			[]string{canned.FakeBucketName, "a", "-o"},
			"Unexpected -o",
		},
	}

	// Run each test case.
	for i, tc := range testCases {
		cmd := exec.Command(t.helperPath)
		cmd.Args = append(cmd.Args, tc.args...)
		cmd.Env = []string{}

		output, err := cmd.CombinedOutput()
		ExpectThat(err, Error(HasSubstr("exit status")), "case %d", i)
		ExpectThat(string(output), MatchesRegexp(tc.expectedOutput), "case %d", i)
	}
}

func (t *MountHelperTest) SuccessfulMount() {
	var err error
	var fi os.FileInfo

	// Mount.
	args := []string{canned.FakeBucketName, t.dir}

	err = t.mount(args)
	AssertEq(nil, err)
	defer unmount(t.dir)

	// Check that the file system is available.
	fi, err = os.Lstat(path.Join(t.dir, canned.TopLevelFile))
	AssertEq(nil, err)
	ExpectEq(os.FileMode(0644), fi.Mode())
	ExpectEq(len(canned.TopLevelFile_Contents), fi.Size())
}

func (t *MountHelperTest) ReadOnlyMode() {
	AssertTrue(false, "TODO")
}

func (t *MountHelperTest) ExtraneousOptions() {
	AssertTrue(false, "TODO")
}

func (t *MountHelperTest) FuseSubtype() {
	AssertTrue(false, "TODO")
}
