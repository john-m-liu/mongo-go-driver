package driver

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"

	"github.com/golang/snappy"
	"go.mongodb.org/mongo-driver/x/mongo/driver/wiremessage"
)

type CompressionOpts struct {
	Compressor       wiremessage.CompressorID
	ZlibLevel        int
	ZstdLevel        int
	UncompressedSize int32
}

func CompressPlayoad(in []byte, opts CompressionOpts) ([]byte, error) {
	switch opts.Compressor {
	case wiremessage.CompressorNoOp:
		return in, nil
	case wiremessage.CompressorSnappy:
		return snappy.Encode(nil, in), nil
	case wiremessage.CompressorZLib:
		var b bytes.Buffer
		w, err := zlib.NewWriterLevel(&b, opts.ZlibLevel)
		if err != nil {
			return nil, err
		}
		_, err = w.Write(in)
		if err != nil {
			return nil, err
		}
		err = w.Close()
		if err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	case wiremessage.CompressorZstd:
		return zstdCompress(in, opts.ZstdLevel)
	default:
		return nil, fmt.Errorf("unknown compressor ID %v", opts.Compressor)
	}
}

func DecompressPayload(in []byte, opts CompressionOpts) ([]byte, error) {
	switch opts.Compressor {
	case wiremessage.CompressorNoOp:
		return in, nil
	case wiremessage.CompressorSnappy:
		uncompressed := make([]byte, opts.UncompressedSize)
		return snappy.Decode(uncompressed, in)
	case wiremessage.CompressorZLib:
		decompressor, err := zlib.NewReader(bytes.NewReader(in))
		if err != nil {
			return nil, err
		}
		uncompressed := make([]byte, opts.UncompressedSize)
		_, err = io.ReadFull(decompressor, uncompressed)
		if err != nil {
			return nil, err
		}
		return uncompressed, nil
	case wiremessage.CompressorZstd:
		return zstdDecompress(in, opts.UncompressedSize)
	default:
		return nil, fmt.Errorf("unknown compressor ID %v", opts.Compressor)
	}
}