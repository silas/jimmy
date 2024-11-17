package migrations

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"

	"buf.build/go/protoyaml"
	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/protobuf/proto"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

var (
	slugifyChars            = regexp.MustCompile(`[^a-z0-9_]`)
	slugifyMultiUnderscores = regexp.MustCompile(`_+`)
	dmlPartitionedPrefix    = regexp.MustCompile(`^(?i)\W*(DELETE|UPDATE)`)
	dmlPrefix               = regexp.MustCompile(`^(?i)\W*INSERT`)
	dmlProtoBundle          = regexp.MustCompile(`^(?i)\W*(CREATE|ALTER)\W+PROTO\W+BUNDLE`)
)

func Ref[T any](v T) *T {
	return &v
}

func Slugify(s string) string {
	s = strings.ToLower(s)
	s = slugifyChars.ReplaceAllString(s, "_")
	s = slugifyMultiUnderscores.ReplaceAllString(s, "_")
	s = strings.Trim(s, "-_")
	return s
}

func Marshal(path string, m proto.Message) error {
	b, err := protoyaml.MarshalOptions{
		Indent: 2,
	}.Marshal(m)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0644)
}

func Unmarshal(path string, m proto.Message) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	validator, err := protovalidate.New()
	if err != nil {
		return err
	}

	return protoyaml.UnmarshalOptions{
		Path:      path,
		Validator: validator,
	}.Unmarshal(b, m)
}

func detectType(sql string) jimmyv1.Type {
	sql = strings.TrimSpace(sql)

	if dmlPrefix.MatchString(sql) {
		return jimmyv1.Type_DML
	} else if dmlPartitionedPrefix.MatchString(sql) {
		return jimmyv1.Type_PARTITIONED_DML
	}

	return jimmyv1.Type_DDL
}

func isProtoDDL(sql string) bool {
	return dmlProtoBundle.MatchString(sql)
}

func checkFile(path, fileType string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%q %s file not found", path, fileType)
		}

		return err
	}

	mode := fileInfo.Mode()

	if mode.IsDir() {
		return fmt.Errorf("%q is a directory, expected a %s file", path, fileType)
	} else if !mode.IsRegular() {
		return fmt.Errorf("%q is not a regular file", path)
	}

	return nil
}
