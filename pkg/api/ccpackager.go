package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/resource"
	"github.com/pkg/errors"

	"fmt"

	"strings"

	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"
)

// Descriptor ...
type Descriptor struct {
	name string
	fqp  string
}

// A list of file extensions that should be packaged into the .tar.gz.
// Files with all other file extenstions will be excluded to minimize the size
// of the install payload.
var keep = []string{".c", ".h", ".s", ".go", ".yaml", ".json"}

var logger = logging.NewLogger("fabsdk/fab")

// CCPackage creates new go lang chaincode package
// To correct instantiation, packager puts file in
// "src/{projectName}" + relative file path
// inside TAR.GZ archive
func CCPackage(chaincodePath, projectName string) (*resource.CCPackage, error) {

	if chaincodePath == "" {
		return nil, errors.New("chaincode path must be provided")
	}

	logger.Debugf("projDir variable=%s", chaincodePath)

	// We generate the tar in two phases: First grab a list of descriptors,
	// and then pack them into an archive.  While the two phases aren't
	// strictly necessary yet, they pave the way for the future where we
	// will need to assemble sources from multiple packages
	descriptors, err := findSource(chaincodePath)
	if err != nil {
		return nil, err
	}

	for i := range descriptors {
		descriptors[i].name = "src/" + projectName + "/" + descriptors[i].name
	}

	tarBytes, err := generateTarGz(descriptors)
	if err != nil {
		return nil, err
	}

	ccPkg := &resource.CCPackage{Type: pb.ChaincodeSpec_GOLANG, Code: tarBytes}

	return ccPkg, nil
}

// -------------------------------------------------------------------------
// findSource(filePath)
// -------------------------------------------------------------------------
// Given an input 'filePath', recursively parse the filesystem for any files
// that fit the criteria for being valid golang source (ISREG + (*.(go|c|h)))
// -------------------------------------------------------------------------
func findSource(filePath string) ([]*Descriptor, error) {
	var descriptors []*Descriptor
	err := filepath.Walk(filePath,
		func(relPath string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fileInfo.Mode().IsRegular() && isSource(relPath) {
				if strings.Contains(relPath, "/META-INF/") {
					relPath = relPath[strings.Index(relPath, "/META-INF/")+1:]
				}
				descriptors = append(descriptors, &Descriptor{name: relPath, fqp: relPath})
			}
			return nil

		})

	return descriptors, err
}

// -------------------------------------------------------------------------
// isSource(path)
// -------------------------------------------------------------------------
// predicate function for determining whether a given path should be
// considered valid source code, based entirely on the extension.  It is
// assumed that other checks for file type have already been performed.
// -------------------------------------------------------------------------
func isSource(filePath string) bool {
	var extension = filepath.Ext(filePath)
	for _, v := range keep {
		if v == extension {
			return true
		}
	}
	return false
}

// -------------------------------------------------------------------------
// generateTarGz(descriptors)
// -------------------------------------------------------------------------
// creates an .tar.gz stream from the provided descriptor entries
// -------------------------------------------------------------------------
func generateTarGz(descriptors []*Descriptor) ([]byte, error) {
	// set up the gzip writer
	var codePackage bytes.Buffer
	gw := gzip.NewWriter(&codePackage)
	tw := tar.NewWriter(gw)
	for _, v := range descriptors {
		logger.Debugf("generateTarGz for %s", v.fqp)
		err := packEntry(tw, gw, v)
		if err != nil {
			err1 := closeStream(tw, gw)
			if err1 != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("packEntry failed and close error %s", err1))
			}
			return nil, errors.Wrap(err, "packEntry failed")
		}
	}
	err := closeStream(tw, gw)
	if err != nil {
		return nil, errors.Wrap(err, "closeStream failed")
	}
	return codePackage.Bytes(), nil

}

func closeStream(tw io.Closer, gw io.Closer) error {
	err := tw.Close()
	if err != nil {
		return err
	}
	err = gw.Close()
	return err
}

func packEntry(tw *tar.Writer, gw *gzip.Writer, descriptor *Descriptor) error {
	file, err := os.Open(descriptor.fqp)
	if err != nil {
		return err
	}
	defer func() {
		err := file.Close()
		if err != nil {
			logger.Warnf("error file close %s", err)
		}
	}()

	if stat, err := file.Stat(); err == nil {

		// now lets create the header as needed for this file within the tarball
		header := new(tar.Header)
		header.Name = descriptor.name
		header.Size = stat.Size()
		header.Mode = int64(stat.Mode())
		// Use a deterministic "zero-time" for all date fields
		header.ModTime = time.Time{}
		header.AccessTime = time.Time{}
		header.ChangeTime = time.Time{}
		// write the header to the tarball archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// copy the file data to the tarball

		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
		if err := tw.Flush(); err != nil {
			return err
		}
		if err := gw.Flush(); err != nil {
			return err
		}

	}
	return nil
}
