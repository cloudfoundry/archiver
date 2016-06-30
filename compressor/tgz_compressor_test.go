package compressor_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/archiver/compressor"
	"code.cloudfoundry.org/archiver/extractor"
)

func retrieveFilePaths(dir string) (results []string) {
	err := filepath.Walk(dir, func(singlePath string, info os.FileInfo, err error) error {
		relative, err := filepath.Rel(dir, singlePath)
		Expect(err).NotTo(HaveOccurred())

		results = append(results, relative)
		return nil
	})

	Expect(err).NotTo(HaveOccurred())

	return results
}

var _ = Describe("Tgz Compressor", func() {
	var compressor Compressor
	var destDir string
	var extracticator extractor.Extractor
	var victimFile *os.File
	var victimDir string

	BeforeEach(func() {
		var err error

		compressor = NewTgz()
		extracticator = extractor.NewDetectable()

		destDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		victimDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		victimFile, err = ioutil.TempFile("", "")
		Expect(err).NotTo(HaveOccurred())

		err = os.Mkdir(filepath.Join(victimDir, "empty"), 0755)
		Expect(err).NotTo(HaveOccurred())

		notEmptyDirPath := filepath.Join(victimDir, "not_empty")

		err = os.Mkdir(notEmptyDirPath, 0755)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(notEmptyDirPath, "some_file"), []byte("stuff"), 0644)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(destDir)
		os.RemoveAll(victimDir)
		os.Remove(victimFile.Name())
	})

	It("compresses the src file to dest file", func() {
		srcFile := victimFile.Name()

		destFile := filepath.Join(destDir, "compress-dst.tgz")

		err := compressor.Compress(srcFile, destFile)
		Expect(err).NotTo(HaveOccurred())

		finalReadingDir, err := ioutil.TempDir(destDir, "final")
		Expect(err).NotTo(HaveOccurred())

		defer os.RemoveAll(finalReadingDir)

		err = extracticator.Extract(destFile, finalReadingDir)
		Expect(err).NotTo(HaveOccurred())

		expectedContent, err := ioutil.ReadFile(srcFile)
		Expect(err).NotTo(HaveOccurred())

		actualContent, err := ioutil.ReadFile(filepath.Join(finalReadingDir, filepath.Base(srcFile)))
		Expect(err).NotTo(HaveOccurred())
		Expect(actualContent).To(Equal(expectedContent))
	})

	It("compresses the src path recursively to dest file", func() {
		srcDir := victimDir

		destFile := filepath.Join(destDir, "compress-dst.tgz")

		err := compressor.Compress(srcDir+"/", destFile)
		Expect(err).NotTo(HaveOccurred())

		finalReadingDir, err := ioutil.TempDir(destDir, "final")
		Expect(err).NotTo(HaveOccurred())

		err = extracticator.Extract(destFile, finalReadingDir)
		Expect(err).NotTo(HaveOccurred())

		expectedFilePaths := retrieveFilePaths(srcDir)
		actualFilePaths := retrieveFilePaths(finalReadingDir)

		Expect(actualFilePaths).To(Equal(expectedFilePaths))

		emptyDir, err := os.Open(filepath.Join(finalReadingDir, "empty"))
		Expect(err).NotTo(HaveOccurred())

		emptyDirInfo, err := emptyDir.Stat()
		Expect(err).NotTo(HaveOccurred())

		Expect(emptyDirInfo.IsDir()).To(BeTrue())
	})
})
