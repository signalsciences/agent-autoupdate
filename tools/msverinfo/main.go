package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/josephspurrier/goversioninfo"
)

const company = "Fastly"
const product = company + " Agent-Autoupdate"
const appname = "Agent-Autoupdate"

type Version struct {
	Major int
	Minor int
	Patch int
}

func main() {
	var version, flagIcon, buildSHA string

	flag.StringVar(&flagIcon, "icon", "", "icon file")
	flag.StringVar(&version, "version", "1.0.0.0", "version")
	flag.StringVar(&buildSHA, "buildsha", "", "build sha")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	v, err := ParseVersion(version)
	if err != nil {
		log.Fatalf("Error reading VERSION %s", err)
	}
	fver := goversioninfo.FileVersion{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	}

	fullver := fmt.Sprintf("%s (build %s)", version, buildSHA)

	vi := &goversioninfo.VersionInfo{
		FixedFileInfo: goversioninfo.FixedFileInfo{
			FileVersion:    fver,
			ProductVersion: fver,
			FileFlagsMask:  "3f", // VS_FFI_FILEFLAGSMASK
			FileFlags:      "00", // None
			FileOS:         "",   // TBD
			FileType:       "01", // VFT_APP
		},
		StringFileInfo: goversioninfo.StringFileInfo{
			Comments:         "", // TBD
			CompanyName:      company,
			FileDescription:  product + " " + fullver,
			FileVersion:      fullver,
			InternalName:     appname,
			LegalCopyright:   time.Now().Format("2006 " + company),
			LegalTrademarks:  "",
			OriginalFilename: appname + ".exe",
			ProductName:      product,
			ProductVersion:   fullver,
		},
		IconPath:     flagIcon,
		ManifestPath: "", // TBD
	}
	vi.Build()
	vi.Walk()
	out := filepath.Clean("./resource.syso")
	if err := vi.WriteSyso(out, "amd64"); err != nil {
		log.Fatalf("Error writing syso %s: %s", out, err)
	}
}

func ParseVersion(s string) (Version, error) {
	// Optional: strip leading "v", e.g. "v1.2.3"
	s = strings.TrimPrefix(s, "v")

	parts := strings.Split(s, ".")
	if len(parts) < 3 {
		return Version{}, fmt.Errorf("invalid version %q: want MAJOR.MINOR.PATCH", s)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version %q: %w", parts[0], err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version %q: %w", parts[1], err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version %q: %w", parts[2], err)
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}
