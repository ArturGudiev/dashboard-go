package repositories

import "testing"

func TestFileAliasPathMatches(t *testing.T) {
	tests := []struct {
		name   string
		stored string
		target string
		want   bool
	}{
		{
			name:   "same unix absolute",
			stored: "/Users/arturgudiev/Data/dashboard_files/tasks/1_foo/note.md",
			target: "/Users/arturgudiev/Data/dashboard_files/tasks/1_foo/note.md",
			want:   true,
		},
		{
			name:   "windows stored vs unix absolute same logical",
			stored: `D:\Data\dashboard_files\tasks\1_foo\note.md`,
			target: "/Users/arturgudiev/Data/dashboard_files/tasks/1_foo/note.md",
			want:   true,
		},
		{
			name:   "windows stored vs logical relative",
			stored: `D:\Data\dashboard_files\tasks\1_foo\note.md`,
			target: "tasks/1_foo/note.md",
			want:   true,
		},
		{
			name:   "bin suffix ignored",
			stored: "/Users/a/Data/dashboard_files/tasks/1_foo/note.md.bin",
			target: "tasks/1_foo/note.md",
			want:   true,
		},
		{
			name:   "case insensitive",
			stored: `D:\Data\Dashboard_Files\Tasks\1_Foo\Note.MD`,
			target: "tasks/1_foo/note.md",
			want:   true,
		},
		{
			name:   "different files",
			stored: "tasks/1_foo/note.md",
			target: "tasks/2_bar/note.md",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fileAliasPathMatches(tt.stored, tt.target); got != tt.want {
				t.Fatalf("fileAliasPathMatches(%q, %q) = %v, want %v", tt.stored, tt.target, got, tt.want)
			}
		})
	}
}
