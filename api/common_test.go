package api

import (
	stderrors "errors"
	"net/http"
	"testing"

	"github.com/pagefaultgames/rogueserver/errors"
	"github.com/stretchr/testify/assert"
)

func TestStatusCodeFromError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		code := statusCodeFromError(nil)
		assert.Equal(t, http.StatusInternalServerError, code)
	})
	t.Run("http error", func(t *testing.T) {
		err := errors.NewHttpError(http.StatusTeapot, "teapot")
		code := statusCodeFromError(err)
		assert.Equal(t, http.StatusTeapot, code)
	})

	t.Run("standard error", func(t *testing.T) {
		err := stderrors.New("standard error")
		code := statusCodeFromError(err)
		assert.Equal(t, http.StatusInternalServerError, code)
	})
}
