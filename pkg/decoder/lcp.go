package decoder

import (
	"bytes"
	"compress/flate"
	"errors"
	"io"
	"io/ioutil"

	"github.com/readium/go-toolkit/pkg/decoder/lcp"
	"github.com/readium/go-toolkit/pkg/pub"
)

func init() {
	decoderList = append(decoderList, List{decoderAlgorithm: "http://www.w3.org/2001/04/xmlenc#aes256-cbc", decoderScheme: "http://readium.org/2014/01/lcp", decoder: DecodeLCP})
}

// DecodeLCP decode lcp encrypted file
func DecodeLCP(publication *pub.Manifest, link pub.Link, reader io.ReadSeeker) (io.ReadSeeker, error) {

	if lcp.HasGoodKey(publication) == false {
		return nil, errors.New(missingOrBadKey)
	}

	cipherRes, errDec := lcp.DecryptData(publication, link, reader)
	if errDec != nil {
		return nil, errDec
	}

	if link.Properties.Encryption.Compression == "deflate" {
		flateReader := flate.NewReader(bytes.NewReader(cipherRes.Bytes()))
		buff, _ := ioutil.ReadAll(flateReader)
		flateReader.Close()
		readerSeeker := bytes.NewReader(buff)

		return readerSeeker, nil
	}

	readerSeeker := bytes.NewReader(cipherRes.Bytes())
	return readerSeeker, nil
}
