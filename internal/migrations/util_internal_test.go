package migrations

import (
	"testing"

	"github.com/stretchr/testify/require"

	jimmyv1 "github.com/silas/jimmy/internal/pb/jimmy/v1"
)

func TestDetectType(t *testing.T) {
	testCases := []struct {
		SQL      string
		Expected jimmyv1.Type
	}{
		{
			SQL:      "",
			Expected: jimmyv1.Type_DDL,
		},
		{
			SQL:      "CREATE TABLE TEST",
			Expected: jimmyv1.Type_DDL,
		},
		{
			SQL:      "CREATE PROTO BUNDLE",
			Expected: jimmyv1.Type_DDL,
		},
		{
			SQL:      "INSERT INTO test",
			Expected: jimmyv1.Type_DML,
		},
		{
			SQL:      "\ninsert \ninto\ntest",
			Expected: jimmyv1.Type_DML,
		},
		{
			SQL:      "UPDATE TABLE test",
			Expected: jimmyv1.Type_PARTITIONED_DML,
		},
		{
			SQL:      "\nupdate\ntable\ntest",
			Expected: jimmyv1.Type_PARTITIONED_DML,
		},
		{
			SQL:      "DELETE FROM TABLE test",
			Expected: jimmyv1.Type_PARTITIONED_DML,
		},
		{
			SQL:      "\ndelete \n from  \ntable \ntest\n",
			Expected: jimmyv1.Type_PARTITIONED_DML,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.SQL, func(t *testing.T) {
			require.Equal(t, tc.Expected, detectType(tc.SQL))
		})
	}
}

func TestIsProtoDDL(t *testing.T) {
	testCases := []struct {
		SQL      string
		Expected bool
	}{
		{
			SQL:      "",
			Expected: false,
		},
		{
			SQL:      "CREATE TABLE TEST",
			Expected: false,
		},
		{
			SQL:      "CREATE PROTO BUNDLE (test)",
			Expected: true,
		},
		{
			SQL:      "\ncreate\n   proto \nbundle (test)",
			Expected: true,
		},
		{
			SQL:      "ALTER PROTO BUNDLE UPDATE (test)",
			Expected: true,
		},
		{
			SQL:      "DROP PROTO BUNDLE",
			Expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.SQL, func(t *testing.T) {
			require.Equal(t, tc.Expected, isProtoDDL(tc.SQL))
		})
	}
}
