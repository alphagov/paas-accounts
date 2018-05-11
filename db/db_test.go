package db_test

import (
	"time"

	. "github.com/alphagov/paas-accounts/db"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DB", func() {
	var (
		db         *DB
		tempDB     *TempDB
		frozenTime time.Time
	)

	BeforeEach(func() {
		frozenTime = time.Date(2001, 1, 1, 1, 1, 1, 0, time.UTC)

		var err error
		tempDB, err = NewTempDB()
		Expect(err).ToNot(HaveOccurred())

		db, err = NewDB(tempDB.TempConnectionString)
		Expect(err).ToNot(HaveOccurred())

		Expect(db.Init()).To(Succeed())
	})

	AfterEach(func() {
		db.Close()
		Expect(tempDB.Close()).To(Succeed())
	})

	It("should run migrations a second time idempotently", func() {
		Expect(db.Init()).To(Succeed())
	})

	Describe("Document", func() {
		It("should put and get a document", func() {
			input := Document{
				Name:      "document",
				Content:   "some agreement terms",
				ValidFrom: frozenTime,
			}

			err := db.PutDocument(input)
			Expect(err).ToNot(HaveOccurred())

			doc, err := db.GetDocument(input.Name)
			Expect(err).ToNot(HaveOccurred())
			Expect(doc).To(Equal(input))
		})

		It("should fail to put a document without a name", func() {
			input := Document{
				Name:      "",
				Content:   "bad-document",
				ValidFrom: frozenTime,
			}

			err := db.PutDocument(input)
			Expect(err).To(MatchError(ContainSubstring("documents_name_check")))
		})

		It("should fail to put a document without content", func() {
			input := Document{
				Name:      "document",
				Content:   "",
				ValidFrom: frozenTime,
			}

			err := db.PutDocument(input)
			Expect(err).To(MatchError(ContainSubstring("documents_content_check")))
		})

		It("should fail to put a document without valid_from", func() {
			input := Document{
				Name:    "document",
				Content: "some-content",
			}

			err := db.PutDocument(input)
			Expect(err).To(MatchError(ContainSubstring("documents_valid_from_check")))
		})

		It("should fail to put a document older than any existing document with the same name", func() {
			doc1 := Document{
				Name:      "document",
				Content:   "some-content",
				ValidFrom: time.Date(2002, 2, 2, 2, 2, 2, 0, time.UTC),
			}
			doc2 := Document{
				Name:      "document",
				Content:   "some-updated-content",
				ValidFrom: time.Date(2001, 1, 1, 1, 1, 1, 0, time.UTC),
			}

			Expect(db.PutDocument(doc1)).To(Succeed())
			Expect(db.PutDocument(doc2)).To(MatchError(ContainSubstring("cannot_alter_document_history")))
		})
	})
})
