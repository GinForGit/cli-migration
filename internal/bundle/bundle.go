// Package bundle handles packing and unpacking of manifest archives.
package bundle

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/GinForGit/cli-migration/internal/manifest"
)

// Metadata describes a bundle archive.
type Metadata struct {
	CliMigVersion    string    `json:"cli_mig_version"`
	CreatedAt        time.Time `json:"created_at"`
	ManifestPath     string    `json:"manifest_path"`
	IncludedConfigs  []string  `json:"included_configs"`
	SourceOS         string    `json:"source_os"`
	SourceArch       string    `json:"source_arch"`
}

// PackOptions controls which extra files are included in a bundle.
type PackOptions struct {
	IncludeConfigs bool
	ConfigPaths    []string
}

// Pack creates a tar.gz bundle from a manifest file.
func Pack(manifestPath, outputPath string, opts PackOptions) error {
	m, err := manifest.Load(manifestPath)
	if err != nil {
		return err
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create bundle: %w", err)
	}
	defer out.Close()

	gz := gzip.NewWriter(out)
	defer gz.Close()

	tw := tar.NewWriter(gz)
	defer tw.Close()

	// Add manifest.
	manifestName := "manifest.yaml"
	if err := addFile(tw, manifestPath, manifestName); err != nil {
		return err
	}

	// Add configs.
	var included []string
	if opts.IncludeConfigs {
		for _, src := range opts.ConfigPaths {
			if _, err := os.Stat(src); err != nil {
				continue
			}
			base := filepath.Base(src)
			dest := filepath.Join("configs", base)
			if err := addFile(tw, src, dest); err != nil {
				return err
			}
			included = append(included, dest)
		}
	}

	// Add metadata.
	meta := Metadata{
		CliMigVersion:   "dev",
		CreatedAt:       time.Now().UTC(),
		ManifestPath:    manifestName,
		IncludedConfigs: included,
		SourceOS:        string(m.Source.OS),
		SourceArch:      m.Source.Arch,
	}
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	if err := addBytes(tw, metaData, "metadata.json", 0o644); err != nil {
		return err
	}

	return nil
}

// Unpack extracts a bundle to a directory and returns the manifest path.
func Unpack(bundlePath, destDir string) (string, error) {
	f, err := os.Open(bundlePath)
	if err != nil {
		return "", fmt.Errorf("open bundle: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("read bundle gzip: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var manifestPath string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read bundle tar: %w", err)
		}

		target := filepath.Join(destDir, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return "", err
			}
			out, err := os.Create(target)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			if hdr.Name == "manifest.yaml" {
				manifestPath = target
			}
		}
	}

	if manifestPath == "" {
		return "", fmt.Errorf("bundle does not contain manifest.yaml")
	}
	return manifestPath, nil
}

func addFile(tw *tar.Writer, srcPath, destName string) error {
	info, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	hdr := &tar.Header{
		Name:    destName,
		Mode:    int64(info.Mode().Perm()),
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	if _, err := io.Copy(tw, f); err != nil {
		return err
	}
	return nil
}

func addBytes(tw *tar.Writer, data []byte, name string, mode int64) error {
	hdr := &tar.Header{
		Name:    name,
		Mode:    mode,
		Size:    int64(len(data)),
		ModTime: time.Now().UTC(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

// DefaultConfigPaths returns common dotfiles to include in a bundle.
func DefaultConfigPaths(home string) []string {
	return []string{
		filepath.Join(home, ".gitconfig"),
		filepath.Join(home, ".ssh", "config"),
	}
}
