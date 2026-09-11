// Copyright 2018 Paul Greenberg (greenpau@outlook.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ovn_exporter

import (
	"slices"
	"testing"
)

func TestNewExporter(t *testing.T) {
	if _, err := NewExporter(Options{}); err != nil {
		t.Errorf("expected no error, but got %q", err)
	}
}

func TestParseComponents(t *testing.T) {
	got, err := ParseComponents("")
	if err != nil || len(got) != 0 {
		t.Fatalf("empty list: got %v, %v; want no components", got, err)
	}

	got, err = ParseComponents(" ovn-northd , ovsdb-server-northbound,ovn-northd ")
	if err != nil {
		t.Fatalf("valid list: unexpected error %v", err)
	}
	if want := []string{"ovn-northd", "ovsdb-server-northbound"}; !slices.Equal(got, want) {
		t.Fatalf("valid list: got %v, want %v", got, want)
	}

	if _, err := ParseComponents("ovs-vswitchd"); err == nil {
		t.Fatal("unsupported component: want an error")
	}
}

func TestEnabledComponents(t *testing.T) {
	e := &Exporter{}
	e.SetComponents([]string{"ovn-northd", "ovsdb-server-northbound"})
	got := e.enabledComponents(DefaultComponents...)
	if want := []string{"ovsdb-server-northbound", "ovn-northd"}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	e.SetComponents(nil)
	if got := e.enabledComponents(DefaultComponents...); len(got) != 0 {
		t.Fatalf("no components selected: got %v", got)
	}
}
