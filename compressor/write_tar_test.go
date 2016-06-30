package compressor_test

import (
	"archive/tar"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/archiver/compressor"
)

var _ = Describe("WriteTar", func() {
	var srcPath string
	var buffer *bytes.Buffer
	var writeErr error

	BeforeEach(func() {
		dir, err := ioutil.TempDir("", "archive-dir")
		Expect(err).NotTo(HaveOccurred())

		err = os.Mkdir(filepath.Join(dir, "outer-dir"), 0755)
		Expect(err).NotTo(HaveOccurred())

		err = os.Mkdir(filepath.Join(dir, "outer-dir", "inner-dir"), 0755)
		Expect(err).NotTo(HaveOccurred())

		innerFile, err := os.Create(filepath.Join(dir, "outer-dir", "inner-dir", "some-file"))
		Expect(err).NotTo(HaveOccurred())

		_, err = innerFile.Write([]byte("sup"))
		Expect(err).NotTo(HaveOccurred())

		err = os.Symlink("some-file", filepath.Join(dir, "outer-dir", "inner-dir", "some-symlink"))
		Expect(err).NotTo(HaveOccurred())

		srcPath = filepath.Join(dir, "outer-dir")
		buffer = new(bytes.Buffer)
	})

	JustBeforeEach(func() {
		writeErr = WriteTar(srcPath, buffer)
	})

	It("returns a reader representing a .tar stream", func() {
		Expect(writeErr).NotTo(HaveOccurred())

		reader := tar.NewReader(buffer)

		header, err := reader.Next()
		Expect(err).NotTo(HaveOccurred())
		Expect(header.Name).To(Equal("outer-dir/"))
		Expect(header.FileInfo().IsDir()).To(BeTrue())

		header, err = reader.Next()
		Expect(err).NotTo(HaveOccurred())
		Expect(header.Name).To(Equal("outer-dir/inner-dir/"))
		Expect(header.FileInfo().IsDir()).To(BeTrue())

		header, err = reader.Next()
		Expect(err).NotTo(HaveOccurred())
		Expect(header.Name).To(Equal("outer-dir/inner-dir/some-file"))
		Expect(header.FileInfo().IsDir()).To(BeFalse())

		contents, err := ioutil.ReadAll(reader)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(contents)).To(Equal("sup"))

		header, err = reader.Next()
		Expect(err).NotTo(HaveOccurred())
		Expect(header.Name).To(Equal("outer-dir/inner-dir/some-symlink"))
		Expect(header.FileInfo().Mode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
		Expect(header.Linkname).To(Equal("some-file"))
	})

	Context("with a trailing slash", func() {
		BeforeEach(func() {
			srcPath = srcPath + "/"
		})

		It("archives the directory's contents", func() {
			Expect(writeErr).NotTo(HaveOccurred())

			reader := tar.NewReader(buffer)

			header, err := reader.Next()
			Expect(err).NotTo(HaveOccurred())
			Expect(header.Name).To(Equal("./"))
			Expect(header.FileInfo().IsDir()).To(BeTrue())

			header, err = reader.Next()
			Expect(err).NotTo(HaveOccurred())
			Expect(header.Name).To(Equal("inner-dir/"))
			Expect(header.FileInfo().IsDir()).To(BeTrue())

			header, err = reader.Next()
			Expect(err).NotTo(HaveOccurred())
			Expect(header.Name).To(Equal("inner-dir/some-file"))
			Expect(header.FileInfo().IsDir()).To(BeFalse())

			contents, err := ioutil.ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("sup"))

			header, err = reader.Next()
			Expect(err).NotTo(HaveOccurred())
			Expect(header.Name).To(Equal("inner-dir/some-symlink"))
			Expect(header.FileInfo().Mode() & os.ModeSymlink).To(Equal(os.ModeSymlink))
			Expect(header.Linkname).To(Equal("some-file"))
		})
	})

	Context("with a single file", func() {
		BeforeEach(func() {
			srcPath = filepath.Join(srcPath, "inner-dir", "some-file")
		})

		It("archives the single file at the root", func() {
			Expect(writeErr).NotTo(HaveOccurred())

			reader := tar.NewReader(buffer)

			header, err := reader.Next()
			Expect(err).NotTo(HaveOccurred())
			Expect(header.Name).To(Equal("some-file"))
			Expect(header.FileInfo().IsDir()).To(BeFalse())

			contents, err := ioutil.ReadAll(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(contents)).To(Equal("sup"))
		})
	})

	Context("when there is no file at the given path", func() {
		BeforeEach(func() {
			srcPath = filepath.Join(srcPath, "barf")
		})

		It("returns an error", func() {
			Expect(writeErr).To(BeAssignableToTypeOf(&os.PathError{}))
		})
	})
})
