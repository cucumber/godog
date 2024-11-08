package godog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttach(t *testing.T) {

	ctx := context.Background()

	ctx = Attach(ctx, Attachment{Body: []byte("body1"), FileName: "fileName1", MediaType: "mediaType1"})
	ctx = Attach(ctx, Attachment{Body: []byte("body2"), FileName: "fileName2", MediaType: "mediaType2"})

	attachments := Attachments(ctx)

	assert.Equal(t, 2, len(attachments))

	assert.Equal(t, []byte("body1"), attachments[0].Body)
	assert.Equal(t, "fileName1", attachments[0].FileName)
	assert.Equal(t, "mediaType1", attachments[0].MediaType)

	assert.Equal(t, []byte("body2"), attachments[1].Body)
	assert.Equal(t, "fileName2", attachments[1].FileName)
	assert.Equal(t, "mediaType2", attachments[1].MediaType)
}
