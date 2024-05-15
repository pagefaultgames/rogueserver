package account

import (
	"net/http"
	"testing"

	"github.com/pagefaultgames/rogueserver/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateUsernamePassword(t *testing.T) {
	t.Run("valid username and password", func(t *testing.T) {
		err := validateUsernamePassword("validUser", "validPass")
		assert.NoError(t, err)
	})

	t.Run("invalid username", func(t *testing.T) {
		err := validateUsernamePassword("", "validPass")
		require.NotNil(t, err)
		assert.Equal(t, err, errors.NewHttpError(http.StatusBadRequest, "invalid username"))
	})

	t.Run("invalid password", func(t *testing.T) {
		err := validateUsernamePassword("validUser", "123")
		require.NotNil(t, err)
		assert.Equal(t, err, errors.NewHttpError(http.StatusBadRequest, "invalid password"))
	})

	t.Run("invalid username and password", func(t *testing.T) {
		err := validateUsernamePassword("", "123")
		require.NotNil(t, err)
		assert.Equal(t, err, errors.NewHttpError(http.StatusBadRequest, "invalid username"))
	})
}

func TestValidateUsername(t *testing.T) {
	t.Run("valid username", func(t *testing.T) {
		err := validateUsername("validUser")
		assert.NoError(t, err)
	})

	t.Run("invalid username", func(t *testing.T) {
		err := validateUsername("")
		require.NotNil(t, err)
		assert.Equal(t, err, errors.NewHttpError(http.StatusBadRequest, "invalid username"))
	})
}

func TestValidatePassword(t *testing.T) {
	t.Run("valid password", func(t *testing.T) {
		err := validatePassword("validPass")
		assert.NoError(t, err)
	})

	t.Run("invalid password", func(t *testing.T) {
		err := validatePassword("123")
		require.NotNil(t, err)
		assert.Equal(t, err, errors.NewHttpError(http.StatusBadRequest, "invalid password"))
	})
}
