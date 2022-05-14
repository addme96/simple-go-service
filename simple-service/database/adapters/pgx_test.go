package adapters_test

import (
	"context"

	"github.com/addme96/simple-go-service/simple-service/database/adapters"
	"github.com/jackc/pgx/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pgx", func() {
	Context("Connect", func() {
		It("calls f(ctx, connString)", func() {
			By("arranging")
			expectedPgxConn := &pgx.Conn{}
			adapter := adapters.Pgx(func(ctx context.Context, connString string) (*pgx.Conn, error) {
				return expectedPgxConn, nil
			})
			ctx := context.Background()
			connString := "connString"

			By("acting")
			pgxConn, err := adapter.Connect(ctx, connString)

			By("asserting")
			Expect(err).NotTo(HaveOccurred())
			Expect(pgxConn).To(Equal(expectedPgxConn))
		})
	})
})
