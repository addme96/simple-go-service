package repositories_test

import (
	"context"
	"errors"
	"regexp"

	"github.com/addme96/simple-go-service/simple-service/entities"
	"github.com/addme96/simple-go-service/simple-service/repositories"
	"github.com/addme96/simple-go-service/simple-service/repositories/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pashagolub/pgxmock"
)

var _ = Describe("Resource", func() {
	var (
		ctrl     *gomock.Controller
		mockDB   *mocks.MockDB
		repo     *repositories.Resource
		ctx      context.Context
		mockConn pgxmock.PgxConnIface
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockDB = mocks.NewMockDB(ctrl)
		repo = repositories.NewResource(mockDB)
		ctx = context.Background()
		mockConn, _ = pgxmock.NewConn()
	})

	expectedErr := errors.New("some error")

	Context("Create", func() {
		query := "INSERT into resources (name) VALUES ($1)"

		Context("happy path", func() {
			It("creates the resource", func() {
				By("arranging")
				resourceToCreate := entities.Resource{ID: 101, Name: "Resource Name"}
				mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
				mockConn.ExpectPrepare("createResource", regexp.QuoteMeta(query)).
					ExpectExec().WithArgs(resourceToCreate.Name).WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mockConn.ExpectClose()

				By("acting")
				err := repo.Create(ctx, resourceToCreate)

				By("asserting")
				Expect(err).NotTo(HaveOccurred())
				Expect(mockConn.ExpectationsWereMet()).To(Succeed())
			})
		})

		Context("not so happy path", func() {
			When("GetConn fails", func() {
				It("returns error", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(nil, expectedErr)

					By("acting")
					err := repo.Create(ctx, entities.Resource{})

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})

			When("Prepare fails", func() {
				It("returns error", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					mockConn.ExpectPrepare("createResource", regexp.QuoteMeta(query)).WillReturnError(expectedErr)
					mockConn.ExpectClose()

					By("acting")
					err := repo.Create(ctx, entities.Resource{})

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})

			When("Exec fails", func() {
				It("returns error", func() {
					By("arranging")
					expectedResource := entities.Resource{ID: 101, Name: "Resource Name"}
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					mockConn.ExpectPrepare("createResource", regexp.QuoteMeta(query)).
						ExpectExec().WithArgs(expectedResource.Name).WillReturnError(expectedErr)
					mockConn.ExpectClose()

					By("acting")
					err := repo.Create(ctx, expectedResource)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})
		})
	})

	Context("Read", func() {
		query := "SELECT id, name FROM resources WHERE id=$1"

		Context("happy path", func() {
			It("reads the resource", func() {
				By("arranging")
				expectedResource := entities.Resource{ID: 101, Name: "Resource Name"}
				mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
				rows := pgxmock.NewRows([]string{"id", "name"}).AddRow(expectedResource.ID, expectedResource.Name)
				mockConn.ExpectPrepare("readResource", regexp.QuoteMeta(query)).ExpectQuery().
					WithArgs(expectedResource.ID).WillReturnRows(rows)
				mockConn.ExpectClose()

				By("acting")
				res, err := repo.Read(ctx, expectedResource.ID)

				By("asserting")
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(&expectedResource))
				Expect(mockConn.ExpectationsWereMet()).To(Succeed())
			})
		})

		Context("not so happy path", func() {
			When("GetConn fails", func() {
				It("returns error", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(nil, expectedErr)

					By("acting")
					res, err := repo.Read(ctx, 101)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(res).To(BeNil())
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})

			When("Prepare fails", func() {
				It("returns error", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					mockConn.ExpectPrepare("readResource", regexp.QuoteMeta(query)).WillReturnError(expectedErr)
					mockConn.ExpectClose()

					By("acting")
					res, err := repo.Read(ctx, 101)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(res).To(BeNil())
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})

			When("QueryRow fails", func() {
				It("returns error", func() {
					By("arranging")
					resourceID := 101
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					mockConn.ExpectPrepare("readResource", regexp.QuoteMeta(query)).
						ExpectQuery().WithArgs(resourceID).WillReturnError(expectedErr)
					mockConn.ExpectClose()

					By("acting")
					res, err := repo.Read(ctx, resourceID)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(res).To(BeNil())
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})
		})
	})

	Context("ReadAll", func() {
		query := "SELECT id, name FROM resources"

		Context("happy path", func() {
			It("reads one resource", func() {
				By("arranging")
				expectedResource := entities.Resource{ID: 101, Name: "Resource Name"}
				mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
				rows := pgxmock.NewRows([]string{"id", "name"}).AddRow(expectedResource.ID, expectedResource.Name)
				mockConn.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(rows)
				mockConn.ExpectClose()

				By("acting")
				res, err := repo.ReadAll(ctx)

				By("asserting")
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal([]entities.Resource{expectedResource}))
				Expect(mockConn.ExpectationsWereMet()).To(Succeed())
			})

			It("reads two resources", func() {
				By("arranging")
				expectedResources := []entities.Resource{
					{ID: 101, Name: "Resource Name 1"},
					{ID: 102, Name: "Resource Name 2"},
				}
				mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
				rows := pgxmock.NewRows([]string{"id", "name"}).
					AddRow(expectedResources[0].ID, expectedResources[0].Name).
					AddRow(expectedResources[1].ID, expectedResources[1].Name)
				mockConn.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(rows)
				mockConn.ExpectClose()

				By("acting")
				res, err := repo.ReadAll(ctx)

				By("asserting")
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal(expectedResources))
				Expect(mockConn.ExpectationsWereMet()).To(Succeed())
			})
		})

		Context("not so happy path", func() {
			When("GetConn fails", func() {
				It("returns error", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(nil, expectedErr)

					By("acting")
					res, err := repo.ReadAll(ctx)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(res).To(BeNil())
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})

			When("Query fails", func() {
				It("returns error other than pgx.ErrNoRows", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					mockConn.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(expectedErr)
					mockConn.ExpectClose()

					By("acting")
					res, err := repo.ReadAll(ctx)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(res).To(BeNil())
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})

				It("returns empty slice in case of pgx.ErrNoRows occurrence", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					mockConn.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(pgx.ErrNoRows)
					mockConn.ExpectClose()

					By("acting")
					res, err := repo.ReadAll(ctx)

					By("asserting")
					Expect(err).To(BeNil())
					Expect(res).To(BeEmpty())
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})

				It("returns error when scan errors", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					expectedResource := entities.Resource{ID: 101, Name: "Resource Name"}
					rows := pgxmock.NewRows([]string{"id", "name"}).AddRow(expectedResource.ID, expectedResource.Name).
						RowError(0, expectedErr)
					mockConn.ExpectQuery(regexp.QuoteMeta(query)).WillReturnRows(rows)
					mockConn.ExpectClose()

					By("acting")
					res, err := repo.ReadAll(ctx)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(res).To(BeNil())
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})
		})
	})

	Context("Update", func() {
		query := "UPDATE resources SET name = $1 WHERE id=$2"
		Context("happy path", func() {
			Context("happy path", func() {
				It("updates the resource", func() {
					By("arranging")
					newResource := entities.Resource{ID: 0, Name: "Resource Name"}
					currentResourceID := 101
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					mockConn.ExpectPrepare("updateResource", regexp.QuoteMeta(query)).
						ExpectExec().WithArgs(newResource.Name, currentResourceID).
						WillReturnResult(pgxmock.NewResult("UPDATE", 1))
					mockConn.ExpectClose()

					By("acting")
					err := repo.Update(ctx, currentResourceID, newResource)

					By("asserting")
					Expect(err).NotTo(HaveOccurred())
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})
		})

		Context("not so happy path", func() {
			When("GetConn fails", func() {
				It("returns error", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(nil, expectedErr)

					By("acting")
					err := repo.Update(ctx, 101, entities.Resource{})

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})

			When("Prepare fails", func() {
				It("returns error", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					mockConn.ExpectPrepare("updateResource", regexp.QuoteMeta(query)).
						WillReturnError(expectedErr)
					mockConn.ExpectClose()

					By("acting")
					err := repo.Update(ctx, 101, entities.Resource{})

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})

			When("Exec fails", func() {
				It("returns error", func() {
					By("arranging")
					newResource := entities.Resource{ID: 0, Name: "Resource Name"}
					currentResourceID := 101
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					mockConn.ExpectPrepare("updateResource", regexp.QuoteMeta(query)).
						ExpectExec().WithArgs(newResource.Name, currentResourceID).WillReturnError(expectedErr)
					mockConn.ExpectClose()

					By("acting")
					err := repo.Update(ctx, currentResourceID, newResource)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})
		})
	})

	Context("Delete", func() {
		query := "DELETE FROM resources WHERE id=$1"

		Context("happy path", func() {
			It("deletes the resource", func() {
				By("arranging")
				resourceID := 101
				mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
				mockConn.ExpectPrepare("deleteResource", regexp.QuoteMeta(query)).
					ExpectExec().WithArgs(resourceID).WillReturnResult(pgxmock.NewResult("DELETE", 1))
				mockConn.ExpectClose()

				By("acting")
				err := repo.Delete(ctx, resourceID)

				By("asserting")
				Expect(err).NotTo(HaveOccurred())
				Expect(mockConn.ExpectationsWereMet()).To(Succeed())
			})
		})

		Context("not so happy path", func() {
			When("GetConn fails", func() {
				It("returns error", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(nil, expectedErr)

					By("acting")
					err := repo.Delete(ctx, 101)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})

			When("Prepare fails", func() {
				It("returns error", func() {
					By("arranging")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					mockConn.ExpectPrepare("deleteResource", regexp.QuoteMeta(query)).
						WillReturnError(expectedErr)
					mockConn.ExpectClose()

					By("acting")
					err := repo.Delete(ctx, 101)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})

			When("Exec fails", func() {
				It("returns error", func() {
					By("arranging")
					resourceID := 101
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					mockConn.ExpectPrepare("deleteResource", regexp.QuoteMeta(query)).
						ExpectExec().WithArgs(resourceID).WillReturnError(expectedErr)
					mockConn.ExpectClose()

					By("acting")
					err := repo.Delete(ctx, resourceID)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})
		})
	})
})
