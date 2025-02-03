package icyproxy

import (
	"bytes"
	"io"
	"testing"
)

func checkIcyBlock(t *testing.T, b []byte, blockSize int, title string) {
	if len(b) != 1+blockSize {
		t.Fatalf("expected %d bytes, got %d", 1+blockSize, len(b))
	}

	if b[0] != byte(blockSize/16) {
		t.Errorf("invalid size byte: expected %x, got %x", blockSize/16, b[0])
	}

	tagStart := string(b[1:14])
	if tagStart != "StreamTitle='" {
		t.Errorf("unexpected metadata prefix: %s", tagStart)
	}

	mdTitle := string(b[14 : 14+len(title)])
	if mdTitle != title {
		t.Errorf("unexpected title, expected %q, got %q", title, mdTitle)
	}

	tagEnd := string(b[14+len(title) : 14+len(title)+2])
	if tagEnd != "';" {
		t.Errorf("unexpected metadata suffix: %s", tagEnd)
	}

	for i := 14 + len(title) + 2; i < len(b); i++ {
		if b[i] != 0 {
			t.Errorf("non-0 byte at index %d", i)
			break
		}
	}
}

func TestMakeIcyBlock(t *testing.T) {
	b := makeIcyBlock("hello")
	checkIcyBlock(t, b, 32, "hello")
}

func checkNoErr(t *testing.T, desc string, err error) {
	if err != nil {
		t.Fatalf("error %s: %s", desc, err)
	}
}

func checkIsAudioData(t *testing.T, desc string, buf []byte, n int) {
	if n != len(buf) {
		t.Fatalf("unexpected read size for %s: expected %d bytes, got %d", desc, n, len(buf))
	}

	for i := 0; i < n; i++ {
		if buf[i] != 0xff {
			t.Fatalf("unexpected byte at index %d for %s: expected 0xff, got %x", i, desc, buf[i])
		}
	}
}

func checkIsMetadata(t *testing.T, desc string, buf []byte, title string) {
	expected := makeIcyBlock(title)
	if len(buf) != len(expected) {
		t.Fatalf("unexpected read size for %s: expected %d bytes, got %d", desc, len(expected), len(buf))
	}

	for i := 0; i < len(buf); i++ {
		if buf[i] != expected[i] {
			t.Fatalf("unexpected byte at index %d for %s: expected %x, got %x", i, desc, expected[i], buf[i])
		}
	}
}

type shortReader struct {
	r        io.Reader
	readSize int
}

func (s *shortReader) Read(p []byte) (int, error) {
	if len(p) > s.readSize {
		p = p[:s.readSize]
	}

	return s.r.Read(p)
}

func TestReader(t *testing.T) {
	audioData := make([]byte, 3*IcyReaderInterval)
	for i := 0; i < len(audioData); i++ {
		audioData[i] = 0xff
	}

	const TestTitle = "test title"
	const MetadataSize = 33

	t.Run("small buffer", func(t *testing.T) {
		r := &IcyReader{AudioData: bytes.NewReader(audioData), title: TestTitle}
		buf := make([]byte, IcyReaderInterval-IcyReaderInterval/4)

		n, err := r.Read(buf)
		checkNoErr(t, "reading first chunk", err)
		checkIsAudioData(t, "first chunk", buf[:n], len(buf))

		n, err = r.Read(buf)
		checkNoErr(t, "reading second chunk", err)

		// Second chunk should be a short read since we'll insert the metadata
		expectedN := IcyReaderInterval - len(buf)
		checkIsAudioData(t, "second chunk", buf[:n], expectedN)

		remaining3rdChunkAudioDataBytes := len(buf) - MetadataSize

		n, err = r.Read(buf)
		checkNoErr(t, "reading first metadata chunk", err)
		checkIsMetadata(t, "first metadata chunk", buf[:MetadataSize], TestTitle)
		checkIsAudioData(t, "third chunk", buf[MetadataSize:], remaining3rdChunkAudioDataBytes)

		n, err = r.Read(buf)
		checkNoErr(t, "reading fourth chunk", err)
		checkIsAudioData(t, "fourth chunk", buf[:n], IcyReaderInterval-remaining3rdChunkAudioDataBytes)
	})

	t.Run("large buffer", func(t *testing.T) {
		r := &IcyReader{AudioData: bytes.NewReader(audioData), title: TestTitle}
		buf := make([]byte, 4*IcyReaderInterval)
		n, err := r.Read(buf)
		checkNoErr(t, "first read", err)
		checkIsAudioData(t, "first chunk", buf[:n], IcyReaderInterval)

		metadataSize := 33
		expectedN := metadataSize + IcyReaderInterval

		n, err = r.Read(buf)
		checkNoErr(t, "second read", err)
		if n != expectedN {
			t.Fatalf("unexpected read size for second chunk: expected %d bytes, got %d", expectedN, n)
		}
		checkIsMetadata(t, "first metadata chunk", buf[:metadataSize], TestTitle)
		checkIsAudioData(t, "second chunk", buf[metadataSize:n], IcyReaderInterval)

		n, err = r.Read(buf)
		checkNoErr(t, "third read", err)
		if n != expectedN {
			t.Fatalf("unexpected read size for second chunk: expected %d bytes, got %d", expectedN, n)
		}
		checkIsMetadata(t, "first metadata chunk", buf[:metadataSize], TestTitle)
		checkIsAudioData(t, "second chunk", buf[metadataSize:n], IcyReaderInterval)
	})

	t.Run("short reads from input", func(t *testing.T) {
		readSize := IcyReaderInterval/3 + 100
		r := &IcyReader{AudioData: &shortReader{r: bytes.NewReader(audioData), readSize: readSize}, title: TestTitle}
		buf := make([]byte, 4*IcyReaderInterval)

		n, err := r.Read(buf)
		checkNoErr(t, "first read", err)
		checkIsAudioData(t, "first chunk", buf[:n], readSize)

		n, err = r.Read(buf)
		checkNoErr(t, "second read", err)
		checkIsAudioData(t, "second chunk", buf[:n], readSize)

		n, err = r.Read(buf)
		checkNoErr(t, "third read", err)

		expectedN := IcyReaderInterval - 2*readSize
		checkIsAudioData(t, "third chunk", buf[:n], expectedN)

		n, err = r.Read(buf)
		checkNoErr(t, "fourth read", err)
		checkIsMetadata(t, "first metadata chunk", buf[:MetadataSize], TestTitle)
		checkIsAudioData(t, "fourth chunk", buf[MetadataSize:n], n-MetadataSize)
	})
}
