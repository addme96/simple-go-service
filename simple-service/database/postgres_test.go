package database_test

import (
	"context"
	"errors"

	"github.com/addme96/simple-go-service/simple-service/database"
	"github.com/addme96/simple-go-service/simple-service/database/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Postgres", func() {
	var ctx context.Context
	var ctrl *gomock.Controller
	var mockPgx *mocks.MockPgx
	var databaseURL = "dbURL"
	var db *database.DB
	BeforeEach(func() {
		ctx = context.Background()
		ctrl = gomock.NewController(GinkgoT())
		mockPgx = mocks.NewMockPgx(ctrl)
		db = database.NewDB(mockPgx, databaseURL)
	})

	AfterEach(func() {
		defer ctrl.Finish()
	})

	Context("GetConn", func() {
		When("happy path", func() {
			It("creates the connection", func() {
				By("arranging")
				mockPgx.EXPECT().Connect(ctx, databaseURL).Times(1).Return(&pgx.Conn{}, nil)

				By("acting")
				conn, err := db.GetConn(ctx)

				By("asserting")
				Expect(err).NotTo(HaveOccurred())
				Expect(conn).NotTo(BeNil())
			})
		})

		When("not so happy path", func() {
			It("creates the connection", func() {
				By("arranging")
				expectedErr := errors.New("some pgx Connect error")
				mockPgx.EXPECT().Connect(ctx, databaseURL).Times(1).Return(nil, expectedErr)

				By("acting")
				conn, err := db.GetConn(ctx)

				By("asserting")
				Expect(err).To(Equal(expectedErr))
				Expect(conn).To(BeNil())
			})
		})
	})
})
