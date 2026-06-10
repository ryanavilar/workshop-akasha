// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"errors"
	"testing"
)

// The fixture loader from upstream (test_fixtures.go) was lost when this
// source tree was vendored for the workshop. These stubs keep the package
// compiling for production builds; tests that need fixtures fail loudly.
var errFixturesUnavailable = errors.New("db fixtures are not available in this vendored copy; restore pkg/db/test_fixtures.go from upstream vikunja to run tests")

// InitFixtures initialize test fixtures for a test database
func InitFixtures(tablenames ...string) error {
	return errFixturesUnavailable
}

// LoadFixtures load all fixtures for a test database
func LoadFixtures() error {
	return errFixturesUnavailable
}

// LoadAndAssertFixtures loads all fixtures defined before and asserts they are correctly loaded
func LoadAndAssertFixtures(t *testing.T) {
	t.Fatal(errFixturesUnavailable)
}
