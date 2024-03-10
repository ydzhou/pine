package pine

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFileModifications(t *testing.T) {
	buffer := Buffer{}

	buffer.New("", logrus.New())

	t.Run("Insert a string", func(t *testing.T) {
		buffer.InsertString("The quick brown fox jumps over the lazy dog")
		assert.True(t, len(buffer.lines) > 0, "buffer lines not empty")
	})

	t.Run("Delete a line", func(t *testing.T) {
		buffer.DeleteLine()
		assert.Equal(t, 0, len(buffer.lines), "buffer lines empty")
	})
}
