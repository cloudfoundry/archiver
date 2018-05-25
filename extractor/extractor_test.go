package extractor_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/archiver/extractor"
	"code.cloudfoundry.org/archiver/extractor/test_helper"
)

var _ = Describe("Extractor", func() {
	var extractor Extractor

	var extractionDest string
	var extractionSrc string

	BeforeEach(func() {
		var err error

		archive, err := ioutil.TempFile("", "extractor-archive")
		Expect(err).NotTo(HaveOccurred())

		extractionDest, err = ioutil.TempDir("", "extracted")
		Expect(err).NotTo(HaveOccurred())

		extractionSrc = archive.Name()

		extractor = NewDetectable()
	})

	AfterEach(func() {
		os.RemoveAll(extractionSrc)
		os.RemoveAll(extractionDest)
	})

	archiveFiles := []test_helper.ArchiveFile{
		{
			Name: "./",
			Dir:  true,
		},
		{
			Name: "./some-file",
			Body: "some-file-contents",
		},
		{
			Name: "./empty-dir/",
			Dir:  true,
		},
		{
			Name: "./nonempty-dir/",
			Dir:  true,
		},
		{
			Name: "./nonempty-dir/file-in-dir",
			Body: "file-in-dir-contents",
		},
		{
			Name: "./legit-exe-not-a-virus.bat",
			Mode: 0644,
			Body: "rm -rf /",
		},
		{
			Name: "./some-symlink",
			Link: "some-file",
			Mode: 0755,
		},
	}

	extractionTest := func() {
		err := extractor.Extract(extractionSrc, extractionDest)
		Expect(err).NotTo(HaveOccurred())

		fileContents, err := ioutil.ReadFile(filepath.Join(extractionDest, "some-file"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(fileContents)).To(Equal("some-file-contents"))

		fileContents, err = ioutil.ReadFile(filepath.Join(extractionDest, "nonempty-dir", "file-in-dir"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(fileContents)).To(Equal("file-in-dir-contents"))

		executable, err := os.Open(filepath.Join(extractionDest, "legit-exe-not-a-virus.bat"))
		Expect(err).NotTo(HaveOccurred())

		executableInfo, err := executable.Stat()
		Expect(err).NotTo(HaveOccurred())
		Expect(executableInfo.Mode()).To(Equal(os.FileMode(0644)))

		emptyDir, err := os.Open(filepath.Join(extractionDest, "empty-dir"))
		Expect(err).NotTo(HaveOccurred())

		emptyDirInfo, err := emptyDir.Stat()
		Expect(err).NotTo(HaveOccurred())

		Expect(emptyDirInfo.IsDir()).To(BeTrue())

		target, err := os.Readlink(filepath.Join(extractionDest, "some-symlink"))
		Expect(err).NotTo(HaveOccurred())
		Expect(target).To(Equal("some-file"))

		symlinkInfo, err := os.Lstat(filepath.Join(extractionDest, "some-symlink"))
		Expect(err).NotTo(HaveOccurred())

		Expect(symlinkInfo.Mode() & 0755).To(Equal(os.FileMode(0755)))
	}

	Context("when the file is a zip archive", func() {
		BeforeEach(func() {
			test_helper.CreateZipArchive(extractionSrc, archiveFiles)
		})

		It("extracts the ZIP's files, generating directories, and honoring file permissions and symlinks", extractionTest)

		Context("with a bad zip archive", func() {
			BeforeEach(func() {
				test_helper.CreateZipArchive(extractionSrc, []test_helper.ArchiveFile{
					{
						Name: "../some-file",
						Body: "file-in-bad-dir-contents",
					},
				})
			})

			It("does not insecurely extract the file outside of the provided destination", func() {
				subdir := filepath.Join(extractionDest, "subdir")
				Expect(os.Mkdir(subdir, 0777)).To(Succeed())
				err := extractor.Extract(extractionSrc, subdir)
				Expect(err).NotTo(HaveOccurred())

				Expect(filepath.Join(extractionDest, "some-file")).NotTo(BeAnExistingFile())
				Expect(filepath.Join(subdir, "some-file")).To(BeAnExistingFile())
			})
		})
	})

	Context("when the file is a tgz archive", func() {
		BeforeEach(func() {
			test_helper.CreateTarGZArchive(extractionSrc, archiveFiles)
		})

		It("extracts the TGZ's files, generating directories, and honoring file permissions and symlinks", extractionTest)

		Context("with a bad tgz archive", func() {
			BeforeEach(func() {
				test_helper.CreateTarGZArchive(extractionSrc, []test_helper.ArchiveFile{
					{
						Name: "../some-file",
						Body: "file-in-bad-dir-contents",
					},
				})
			})

			It("does not insecurely extract the file outside of the provided destination", func() {
				subdir := filepath.Join(extractionDest, "subdir")
				Expect(os.Mkdir(subdir, 0777)).To(Succeed())
				err := extractor.Extract(extractionSrc, subdir)
				Expect(err).NotTo(HaveOccurred())
				Expect(filepath.Join(extractionDest, "some-file")).NotTo(BeAnExistingFile())
				Expect(filepath.Join(subdir, "some-file")).To(BeAnExistingFile())
			})
		})
	})

	Context("when the file is a tar archive", func() {
		BeforeEach(func() {
			extractor = NewTar()
			test_helper.CreateTarArchive(extractionSrc, archiveFiles)
		})

		It("extracts the TAR's files, generating directories, and honoring file permissions and symlinks", extractionTest)

		Context("with a bad tar archive", func() {
			BeforeEach(func() {
				test_helper.CreateTarArchive(extractionSrc, []test_helper.ArchiveFile{
					{
						Name: "../some-file",
						Body: "file-in-bad-dir-contents",
					},
				})
			})

			It("does not insecurely extract the file outside of the provided destination", func() {
				subdir := filepath.Join(extractionDest, "subdir")
				Expect(os.Mkdir(subdir, 0777)).To(Succeed())
				err := extractor.Extract(extractionSrc, subdir)
				Expect(err).NotTo(HaveOccurred())
				Expect(filepath.Join(extractionDest, "some-file")).NotTo(BeAnExistingFile())
				Expect(filepath.Join(subdir, "some-file")).To(BeAnExistingFile())
			})
		})
	})
})
