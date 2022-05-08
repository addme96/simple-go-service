package repositories_test

import (
	"context"
	"errors"
	"regexp"

	"github.com/addme96/simple-go-service/simple-service/entities"
	"github.com/addme96/simple-go-service/simple-service/repositories"
	"github.com/addme96/simple-go-service/simple-service/repositories/mocks"
	"github.com/golang/mock/gomock"
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

	AfterEach(func() {
		defer ctrl.Finish()
	})
	Context("Create", func() {
		query := "INSERT into resources (name) VALUES ($1)"

		Context("happy path", func() {
			It("creates the resource", func() {
				By("arranging")
				expectedResource := entities.Resource{ID: 101, Name: "Resource Name"}
				mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
				mockConn.ExpectPrepare("createResource", regexp.QuoteMeta(query)).
					ExpectExec().WithArgs(expectedResource.Name).WillReturnResult(pgxmock.NewResult("INSERT", 1))
				mockConn.ExpectClose()

				By("acting")
				err := repo.Create(ctx, expectedResource)

				By("asserting")
				Expect(err).NotTo(HaveOccurred())
				Expect(mockConn.ExpectationsWereMet()).To(Succeed())
			})
		})

		Context("not so happy path", func() {
			When("GetConn fails", func() {
				It("returns error", func() {
					By("arranging")
					expectedResource := entities.Resource{ID: 101, Name: "Resource Name"}
					expectedErr := errors.New("some error")
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(nil, expectedErr)

					By("acting")
					err := repo.Create(ctx, expectedResource)

					By("asserting")
					Expect(err).To(Equal(expectedErr))
					Expect(mockConn.ExpectationsWereMet()).To(Succeed())
				})
			})

			When("Prepare fails", func() {
				It("returns error", func() {
					By("arranging")
					expectedResource := entities.Resource{ID: 101, Name: "Resource Name"}
					mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
					expectedErr := errors.New("prepare error")
					mockConn.ExpectPrepare("createResource", regexp.QuoteMeta(query)).WillReturnError(expectedErr)
					mockConn.ExpectClose()
					By("acting")
					err := repo.Create(ctx, expectedResource)

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
					expectedErr := errors.New("prepare error")
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
})
