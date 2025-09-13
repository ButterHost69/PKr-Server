package handlers

import (
	"context"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/ButterHost69/PKr-Base/pb"
	"github.com/ButterHost69/PKr-Server/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var s CliServiceServer

// ---------- Helper Methods ----------
func setupTestDB(t *testing.T) {
	t.Helper()

	err := db.InitSQLiteDatabase(":memory:")
	require.NoError(t, err)
}

func TestRegister(t *testing.T) {
	setupTestDB(t)

	testcases := []struct {
		test_name    string
		username     string
		password     string
		want_err     bool
		err_contains string
	}{
		{"Valid Testcase 1", "User1", "User1", false, ""},
		{"Valid Testcase 2", "User2", "User2", false, ""},
		{"Valid Testcase 3", "User3", "User1", false, ""},
		{"Duplicate Username 1", "User2", "User234", true, "username is already taken"},
		{"Duplicate Username 2", "User3", "User1234", true, "username is already taken"},
		{"Non-Alphanumeric Username 1", "User ", "User2343", true, "username must be alphanumeric"},
		{"Non-Alphanumeric Username 2", "User#1", "User1234", true, "username must be alphanumeric"},
		{"Non-Alphanumeric Username 3", "User1_", "User1234", true, "username must be alphanumeric"},
		{"Empty Username", "", "User1234", true, "username must be alphanumeric"},
	}

	for _, tc := range testcases {
		t.Run(tc.test_name, func(t *testing.T) {
			// No need to log errors from funcs during tests when we expect errors
			if tc.want_err {
				log.SetOutput(io.Discard)
				defer log.SetOutput(os.Stderr)
			}

			req := &pb.RegisterRequest{
				Username: tc.username, Password: tc.password,
			}

			_, err := s.Register(context.Background(), req)
			if tc.want_err {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), tc.err_contains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// func TestRegisterWorkspace(t *testing.T) {
// 	setupTestDB(t)

// 	// testcases := []struct{}{}
// }
