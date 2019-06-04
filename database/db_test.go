package database_test

import (
	"time"

	. "github.com/alphagov/paas-accounts/database"

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

		It("should not update a document if the content matches the latest version", func() {
			firstDate := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			secondDate := time.Date(2002, 2, 2, 2, 2, 2, 2, time.UTC)
			doc1 := Document{
				Name:      "document",
				Content:   "some-content",
				ValidFrom: firstDate,
			}
			doc2 := Document{
				Name:      "document",
				Content:   "some-content",
				ValidFrom: secondDate,
			}

			Expect(db.PutDocument(doc1)).To(Succeed())
			Expect(db.PutDocument(doc2)).To(Succeed())

			latestVersion, err := db.GetDocument("document")
			Expect(err).NotTo(HaveOccurred())
			Expect(latestVersion.ValidFrom).To(Equal(firstDate))
		})
	})

	Describe("User", func() {
		It("should post a user idempotently", func() {
			user := User{
				UUID:  "00000000-0000-0000-0000-000000000001",
				Email: "example@example.com",
			}

			Expect(db.PostUser(user)).To(Succeed())
			Expect(db.PostUser(user)).To(Succeed())
		})

		It("should fail to post a user without a uuid", func() {
			user := User{
				UUID:  "",
				Email: "example@example.com",
			}

			err := db.PostUser(user)
			Expect(err).To(MatchError(ContainSubstring("invalid input syntax for uuid")))
		})

		It("should update user", func() {
			user := User{
				UUID:  "00000000-0000-0000-0000-000000000001",
				Email: "newexample@example.com",
			}

			Expect(db.PatchUser(user)).To(Succeed())
		})

		It("should return all users", func() {
			user := User{
				UUID:  "00000000-0000-0000-0000-000000000001",
				Email: "example@example.com",
			}
			Expect(db.PostUser(user)).To(Succeed())

			user1 := User{
				UUID:  "00000000-0000-0000-0000-000000000002",
				Email: "newexample@example.com",
			}
			Expect(db.PostUser(user1)).To(Succeed())

			userlist := []string{"00000000-0000-0000-0000-000000000001", "00000000-0000-0000-0000-000000000002"}
			users, err := db.GetUsersByUUID(userlist)
			Expect(err).ToNot(HaveOccurred())
			Expect(users).To(Equal([]User{
				{
					UUID:  user.UUID,
					Email: user.Email,
				},
				{
					UUID:  user1.UUID,
					Email: user1.Email,
				},
			}))
		})

	})

	Describe("Agreement", func() {

		var (
			user     User
			document Document
		)

		BeforeEach(func() {
			user = User{
				UUID:  "00000000-0000-0000-0000-000000000001",
				Email: "example@example.com",
			}
			document = Document{
				Name:      "document",
				Content:   "some agreement terms",
				ValidFrom: frozenTime,
			}
			Expect(db.PostUser(user)).To(Succeed())
			Expect(db.PutDocument(document)).To(Succeed())
		})

		It("should put Agreement", func() {
			agreement := Agreement{
				UserUUID:     user.UUID,
				DocumentName: document.Name,
				Date:         time.Date(2002, 2, 2, 2, 2, 2, 0, time.UTC),
			}

			Expect(db.PutAgreement(agreement)).To(Succeed())

			agreements, err := db.GetAgreementsForUserUUID(user.UUID)
			Expect(err).ToNot(HaveOccurred())
			Expect(agreements).To(HaveLen(1))
			Expect(agreements[0]).To(Equal(agreement))
		})

		It("should fail to put Agreement without a valid user UUID", func() {
			agreement := Agreement{
				UserUUID:     "00000000-0000-0000-0000-000000000002",
				DocumentName: document.Name,
				Date:         time.Date(2002, 2, 2, 2, 2, 2, 0, time.UTC),
			}

			err := db.PutAgreement(agreement)
			Expect(err).To(MatchError(ContainSubstring("agreements_user_uuid_fkey")))
		})

		It("should fail to put Agreement without a valid document name", func() {
			agreement := Agreement{
				UserUUID:     user.UUID,
				DocumentName: "non-existant-doc",
				Date:         time.Date(2002, 2, 2, 2, 2, 2, 0, time.UTC),
			}

			err := db.PutAgreement(agreement)
			Expect(err).To(MatchError(ContainSubstring("agreements_document_not_exist")))
		})

		It("should fail to put Agreement before a document is valid", func() {
			agreement := Agreement{
				UserUUID:     user.UUID,
				DocumentName: "document",
				Date:         time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC),
			}

			err := db.PutAgreement(agreement)
			Expect(err).To(MatchError(ContainSubstring("agreements_document_not_exist")))
		})

		It("should fail to put Agreement without a date", func() {
			agreement := Agreement{
				UserUUID:     user.UUID,
				DocumentName: document.Name,
			}

			err := db.PutAgreement(agreement)
			Expect(err).To(MatchError(ContainSubstring("agreements_date_check")))
		})

	})

	Describe("GetDocumentsForUserUUID", func() {
		var (
			user                                                 User
			documentSuperseded, documentAgreed, documentUnagreed Document
			agreement, agreementSuperseded                       Agreement
		)

		BeforeEach(func() {
			user = User{
				UUID:  "00000000-0000-0000-0000-000000000001",
				Email: "example@example.com",
			}
			Expect(db.PostUser(user)).To(Succeed())

			documentAgreed = Document{
				Name:      "document-agreed",
				Content:   "content agreed",
				ValidFrom: time.Date(2002, 2, 2, 2, 2, 2, 0, time.UTC),
			}

			documentSuperseded = documentAgreed
			documentSuperseded.ValidFrom = documentAgreed.ValidFrom.AddDate(-1, 0, 0)
			documentSuperseded.Content = documentAgreed.Content + " v2"
			Expect(db.PutDocument(documentSuperseded)).To(Succeed())

			agreementSuperseded = Agreement{
				UserUUID:     user.UUID,
				DocumentName: documentSuperseded.Name,
				Date:         documentSuperseded.ValidFrom,
			}
			Expect(db.PutAgreement(agreementSuperseded)).To(Succeed())
			Expect(db.PutDocument(documentAgreed)).To(Succeed())

			agreement = Agreement{
				UserUUID:     user.UUID,
				DocumentName: documentAgreed.Name,
				Date:         time.Date(2004, 4, 4, 4, 4, 4, 0, time.UTC),
			}
			Expect(db.PutAgreement(agreement)).To(Succeed())

			documentUnagreed = Document{
				Name:      "document-unagreed",
				Content:   "content unagreed",
				ValidFrom: time.Date(2003, 3, 3, 3, 3, 3, 0, time.UTC),
			}
			Expect(db.PutDocument(documentUnagreed)).To(Succeed())
		})

		It("should return all relevant documents for a user", func() {
			userDocuments, err := db.GetDocumentsForUserUUID(user.UUID)
			Expect(err).ToNot(HaveOccurred())
			Expect(userDocuments).To(Equal([]UserDocument{
				{
					Name:          documentSuperseded.Name,
					Content:       documentSuperseded.Content,
					ValidFrom:     documentSuperseded.ValidFrom,
					AgreementDate: &agreementSuperseded.Date,
				},
				{
					Name:          documentAgreed.Name,
					Content:       documentAgreed.Content,
					ValidFrom:     documentAgreed.ValidFrom,
					AgreementDate: &agreement.Date,
				},
				{
					Name:          documentUnagreed.Name,
					Content:       documentUnagreed.Content,
					ValidFrom:     documentUnagreed.ValidFrom,
					AgreementDate: nil,
				},
			}))
		})
	})
})
