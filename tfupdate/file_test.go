package tfupdate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func TestUpdateFileExist(t *testing.T) {
	cases := []struct {
		filename string
		src      string
		o        Option
		want     string
		ok       bool
	}{
		{
			filename: "valid.tf",
			src: `
terraform {
  required_version = "0.12.6"
}
`,
			o: Option{
				updateType: "terraform",
				target:     "0.12.7",
			},
			want: `
terraform {
  required_version = "0.12.7"
}
`,
			ok: true,
		},
		{
			filename: "invalid.tf",
			src: `
terraform {
  required_version = "0.12.6"
}
`,
			o: Option{
				updateType: "hoge",
				target:     "0.12.7",
			},
			want: `
terraform {
  required_version = "0.12.6"
}
`,
			ok: false,
		},
		{
			filename: "unformatted_match.tf",
			src: `
terraform {
required_version = "0.12.6"
}
`,
			o: Option{
				updateType: "terraform",
				target:     "0.12.7",
			},
			want: `
terraform {
  required_version = "0.12.7"
}
`,
			ok: true,
		},
		{
			filename: "unformatted_mo_match.tf",
			src: `
terraform {
required_version = "0.12.6"
}
`,
			o: Option{
				updateType: "provider",
				target:     "aws@2.23.0",
			},
			want: `
terraform {
required_version = "0.12.6"
}
`,
			ok: true,
		},
	}
	for _, tc := range cases {
		fs := afero.NewMemMapFs()
		err := afero.WriteFile(fs, tc.filename, []byte(tc.src), 0644)
		if err != nil {
			t.Fatalf("failed to write file: %s", err)
		}

		err = UpdateFile(fs, tc.filename, tc.o)
		if tc.ok && err != nil {
			t.Errorf("UpdateFile() with filename = %s, o = %#v returns unexpected err: %+v", tc.filename, tc.o, err)
		}

		if !tc.ok && err == nil {
			t.Errorf("UpdateFile() with filename = %s, o = %#v expects to return an error, but no error", tc.filename, tc.o)
		}

		got, err := afero.ReadFile(fs, tc.filename)
		if err != nil {
			t.Fatalf("failed to read updated file: %s", err)
		}

		if string(got) != tc.want {
			t.Errorf("UpdateFile() with filename = %s, o = %#v returns %s, but want = %s", tc.filename, tc.o, string(got), tc.want)
		}
	}
}

func TestUpdateFileNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	filename := "not_found.tf"
	o := Option{}

	err := UpdateFile(fs, filename, o)

	if err == nil {
		t.Errorf("UpdateFile() with filename = %s, o = %#v expects to return an error, but no error", filename, o)
	}
}

func TestUpdateDirExist(t *testing.T) {
	cases := []struct {
		rootdir   string
		subdir    string
		filename1 string
		src1      string
		filename2 string
		src2      string
		o         Option
		checkdir  string
		recursive bool
		want1     string
		want2     string
	}{
		{
			rootdir:   "a",
			subdir:    "b",
			filename1: "terraform.tf",
			src1: `
terraform {
  required_version = "0.12.6"
}
`,
			filename2: "provider.tf",
			src2: `
provider "aws" {
  version = "2.11.0"
}
`,
			checkdir:  "a/b",
			recursive: false,
			o: Option{
				updateType: "terraform",
				target:     "0.12.7",
			},
			want1: `
terraform {
  required_version = "0.12.7"
}
`,
			want2: `
provider "aws" {
  version = "2.11.0"
}
`,
		},
		{
			rootdir:   "a",
			subdir:    "b",
			filename1: "terraform.tf",
			src1: `
terraform {
  required_version = "0.12.6"
}
`,
			filename2: "provider.tf",
			src2: `
provider "aws" {
  version = "2.11.0"
}
`,
			checkdir:  "a",
			recursive: true,
			o: Option{
				updateType: "terraform",
				target:     "0.12.7",
			},
			want1: `
terraform {
  required_version = "0.12.7"
}
`,
			want2: `
provider "aws" {
  version = "2.11.0"
}
`,
		},
		{
			rootdir:   "a",
			subdir:    "b",
			filename1: "terraform.tf",
			src1: `
terraform {
  required_version = "0.12.6"
}
`,
			filename2: "provider.tf",
			src2: `
provider "aws" {
  version = "2.11.0"
}
`,
			checkdir:  "a",
			recursive: false,
			o: Option{
				updateType: "terraform",
				target:     "0.12.7",
			},
			want1: `
terraform {
  required_version = "0.12.6"
}
`,
			want2: `
provider "aws" {
  version = "2.11.0"
}
`,
		},
		{
			rootdir:   "a",
			subdir:    ".terraform",
			filename1: "terraform.tf",
			src1: `
terraform {
  required_version = "0.12.6"
}
`,
			filename2: "provider.tf",
			src2: `
provider "aws" {
  version = "2.11.0"
}
`,
			checkdir:  "a",
			recursive: true,
			o: Option{
				updateType: "terraform",
				target:     "0.12.7",
			},
			want1: `
terraform {
  required_version = "0.12.6"
}
`,
			want2: `
provider "aws" {
  version = "2.11.0"
}
`,
		},
		{
			rootdir:   "a",
			subdir:    "b",
			filename1: "terraform.hcl",
			src1: `
terraform {
  required_version = "0.12.6"
}
`,
			filename2: "provider.tf",
			src2: `
provider "aws" {
  version = "2.11.0"
}
`,
			checkdir:  "a/b",
			recursive: false,
			o: Option{
				updateType: "terraform",
				target:     "0.12.7",
			},
			want1: `
terraform {
  required_version = "0.12.6"
}
`,
			want2: `
provider "aws" {
  version = "2.11.0"
}
`,
		},
	}

	for _, tc := range cases {
		fs := afero.NewMemMapFs()
		dirname := filepath.Join(tc.rootdir, tc.subdir)
		err := fs.MkdirAll(dirname, os.ModePerm)
		if err != nil {
			t.Fatalf("failed to create dir: %s", err)
		}

		err = afero.WriteFile(fs, filepath.Join(dirname, tc.filename1), []byte(tc.src1), 0644)
		if err != nil {
			t.Fatalf("failed to write file: %s", err)
		}

		err = afero.WriteFile(fs, filepath.Join(dirname, tc.filename2), []byte(tc.src2), 0644)
		if err != nil {
			t.Fatalf("failed to write file: %s", err)
		}

		err = UpdateDir(fs, tc.checkdir, tc.recursive, tc.o)

		if err != nil {
			t.Errorf("UpdateDir() with dirname = %s, recursive = %t, o = %#v returns an unexpected error: %+v", tc.checkdir, tc.recursive, tc.o, err)
		}

		got1, err := afero.ReadFile(fs, filepath.Join(dirname, tc.filename1))
		if err != nil {
			t.Fatalf("failed to read file: %s", err)
		}

		if string(got1) != tc.want1 {
			t.Errorf("UpdateDir() with dirname = %s, recursive = %t, o = %#v returns %s, but want = %s", dirname, tc.recursive, tc.o, string(got1), tc.want1)
		}

		got2, err := afero.ReadFile(fs, filepath.Join(dirname, tc.filename2))
		if err != nil {
			t.Fatalf("failed to read file: %s", err)
		}

		if string(got2) != tc.want2 {
			t.Errorf("UpdateDir() with dirname = %s, recursive = %t, o = %#v returns %s, but want = %s", dirname, tc.recursive, tc.o, string(got2), tc.want2)
		}
	}
}

func TestUpdateDirNotFound(t *testing.T) {
	fs := afero.NewMemMapFs()
	dirname := "not_found"
	recursive := false
	o := Option{}

	err := UpdateDir(fs, dirname, recursive, o)

	if err == nil {
		t.Errorf("UpdateDir() with dirname = %s, recursive = %t, o = %#v expects to return an error, but no error", dirname, recursive, o)
	}
}