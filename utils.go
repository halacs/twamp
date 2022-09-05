package twamp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

func ReadFromSocket(reader io.Reader, size int) (bytes.Buffer, error) {
	buf := make([]byte, size)
	buffer := *bytes.NewBuffer(buf)
	bytesRead, err := reader.Read(buf)

	if err != nil && bytesRead < size {
		return buffer, errors.New(fmt.Sprintf("readFromSocket: expected %d bytes, got %d", size, bytesRead))
	}

	return buffer, err
}
