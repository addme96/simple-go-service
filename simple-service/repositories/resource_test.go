package repositories_test

import (
	"context"
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
		defer func(mockConn pgxmock.PgxConnIface, ctx context.Context) {
			_ = mockConn.Close(ctx)
			//if err != nil {
			//	panic(err)
			//}
		}(mockConn, ctx)
	})
	Context("Create", func() {
		Context("happy path", func() {
			It("creates the resource", func() {
				By("arranging")
				expectedResource := entities.Resource{ID: 101, Name: "Resource Name"}
				mockDB.EXPECT().GetConn(ctx).Times(1).Return(mockConn, nil)
				query := "INSERT into resources (name) VALUES ($1)"
				mockConn.ExpectPrepare("createResource", regexp.QuoteMeta(query)).
					ExpectExec().WithArgs(expectedResource.Name).WillReturnResult(pgxmock.NewResult("INSERT", 1))

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

					By("acting")

					By("asserting")

				})
			})

			When("Prepare fails", func() {
				It("returns error", func() {
					By("arranging")

					By("acting")

					By("asserting")

				})
			})

			When("Query fails", func() {
				It("returns error", func() {
					By("arranging")

					By("acting")

					By("asserting")

				})
			})
		})
	})
})
