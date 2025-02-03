package icyproxy

import (
	"bytes"
	"io"
	"sync"
)

const IcyReaderInterval = 16000 // bytes

type IcyReader struct {
	mu              sync.Mutex
	AudioData       io.Reader
	title           string
	audioBytesSent  int
	pendingMetadata []byte
}

var _ io.Reader = &IcyReader{}

func makeIcyBlock(streamTitle string) []byte {
	// metadata block: 1 byte for size, then StreamTitle='title';
	// len("StreamTitle='';") == 15
	const BaseSize = 15
	const MaxTitleSize = 255*16 - BaseSize

	if len(streamTitle) > MaxTitleSize {
		streamTitle = streamTitle[:MaxTitleSize]
	}

	blockSize := 16 * (1 + (len(streamTitle)+BaseSize)/16)

	var res bytes.Buffer
	res.Grow(1 + blockSize)
	res.WriteByte(byte(blockSize / 16))
	res.WriteString("StreamTitle='")
	res.WriteString(streamTitle)
	res.WriteString("';")

	pad := 1 + blockSize - res.Len()

	for i := 0; i < pad; i++ {
		res.WriteByte(0)
	}

	return res.Bytes()
}

func (r *IcyReader) Read(buf []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	metadataBytesSent := 0

	if r.pendingMetadata != nil {
		end := len(r.pendingMetadata)
		if end > len(buf) {
			end = len(buf)
		}

		copy(buf, r.pendingMetadata[:end])
		buf = buf[end:]

		if end == len(r.pendingMetadata) {
			// could send all the metadata
			metadataBytesSent = len(r.pendingMetadata)
			r.pendingMetadata = nil
			r.audioBytesSent = 0
		} else {
			// will send the rest of the metadata later
			return len(buf), nil
		}
	}

	bytesUntilNextMetadata := IcyReaderInterval - r.audioBytesSent

	if len(buf) >= bytesUntilNextMetadata {
		buf = buf[:bytesUntilNextMetadata]
	}

	n, err := r.AudioData.Read(buf)
	r.audioBytesSent += n

	if r.audioBytesSent == IcyReaderInterval {
		r.pendingMetadata = makeIcyBlock(r.title)
	}

	return n + metadataBytesSent, err
}

func (r *IcyReader) SetTitle(title string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.title = title
}
