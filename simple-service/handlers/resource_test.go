package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strconv"

	"github.com/addme96/simple-go-service/simple-service/entities"
	"github.com/addme96/simple-go-service/simple-service/handlers"
	"github.com/addme96/simple-go-service/simple-service/handlers/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Resource", func() {
	var mockCtrl *gomock.Controller
	var mockRepo *mocks.MockResourceRepository
	var server *ghttp.Server
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockRepo = mocks.NewMockResourceRepository(mockCtrl)
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		mockCtrl.Finish()
		server.Close()
	})

	Context("NewResource", func() {
		It("creates resource handler with a given Repository", func() {
			handler := handlers.NewResource(mockRepo)
			Expect(*handler).To(Equal(handlers.Resource{Repository: mockRepo}))
		})
	})

	Context("Post", func() {
		BeforeEach(func() {
			server.AppendHandlers(handlers.NewResource(mockRepo).Post)
		})
		When("valid request", func() {
			It("creates the resource", func() {
				By("arranging")
				r := entities.Resource{
					Name: "Resource Name",
				}
				mockRepo.EXPECT().Create(gomock.Any(), r).Times(1).Return(nil)
				body, err := json.Marshal(r)
				Expect(err).ShouldNot(HaveOccurred())

				By("acting")
				resp, err := http.Post(server.URL(), "application/json", bytes.NewReader(body))

				By("asserting")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				respBody, err := io.ReadAll(resp.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(respBody).To(BeEmpty())
			})
		})

		When("invalid Content-Type", func() {
			It("returns 400 Bad Request", func() {
				By("arranging")
				r := entities.Resource{
					Name: "Resource Name",
				}
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
				body, err := json.Marshal(r)
				Expect(err).ShouldNot(HaveOccurred())

				By("acting")
				resp, err := http.Post(server.URL(), "", bytes.NewReader(body))

				By("asserting")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				respBody, err := io.ReadAll(resp.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(respBody).To(BeEmpty())
			})
		})
	})

	Context("GetCtx", func() {
		var resource *entities.Resource
		BeforeEach(func() {
			resourceHandler := handlers.NewResource(mockRepo)
			nextHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				resource = request.Context().Value("resource").(*entities.Resource)
			})
			router := chi.NewRouter()
			router.Route("/resources/{resourceID}", func(r chi.Router) {
				r.Use(resourceHandler.GetCtx)
				r.Get("/", nextHandler.ServeHTTP)
			})
			server.AppendHandlers(router.ServeHTTP)
		})
		AfterEach(func() {
			By("resetting var for 'not found' test cases")
			resource = nil
		})
		When("valid request", func() {
			It("fetches the resource and puts it into the context", func() {
				By("arranging")
				resourceID := 123
				expectedResource := &entities.Resource{
					ID:   resourceID,
					Name: "Resource Name",
				}
				mockRepo.EXPECT().Read(gomock.Any(), resourceID).Times(1).Return(expectedResource, nil)

				By("acting")
				resp, err := http.Get(server.URL() + path.Join("/resources", strconv.Itoa(resourceID)))

				By("asserting")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resource).To(Equal(expectedResource))
			})
			It("returns 404 if resource not found", func() {
				By("arranging")
				resourceID := 123
				mockRepo.EXPECT().Read(gomock.Any(), resourceID).Times(1).Return(nil, pgx.ErrNoRows)

				By("acting")
				resp, err := http.Get(server.URL() + path.Join("/resources", strconv.Itoa(resourceID)))

				By("asserting")
				Expect(err).ToNot(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(resource).To(BeNil())
			})
		})
		When("invalid request", func() {
			It("returns 400 when not int", func() {
				By("arranging")
				mockRepo.EXPECT().Read(gomock.Any(), gomock.Any()).Times(0)

				By("acting")
				resp, err := http.Get(server.URL() + path.Join("/resources", "definitely-not-int"))

				By("asserting")
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(resource).To(BeNil())
			})
		})
	})

	Context("Get", func() {
		When("valid request", func() {
			It("gets the resource if it is in the context", func() {
				By("arranging")
				resourceID := 123
				resource := &entities.Resource{
					ID:   resourceID,
					Name: "Resource Name",
				}
				w := httptest.NewRecorder()
				ctxWithResource := context.WithValue(context.TODO(), "resource", resource)
				req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctxWithResource)
				expectedBody, err := json.Marshal(resource)
				Expect(err).NotTo(HaveOccurred())

				By("acting")
				handlers.NewResource(mockRepo).Get(w, req)
				res := w.Result()
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)

				By("asserting")
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(err).NotTo(HaveOccurred())
				Expect(body).To(MatchJSON(expectedBody))
			})

			It("returns 400 if there is no value in the context", func() {
				By("arranging")
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/", nil)

				By("acting")
				handlers.NewResource(mockRepo).Get(w, req)
				res := w.Result()
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				By("asserting")
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(err).NotTo(HaveOccurred())
				Expect(body).To(BeEmpty())
			})
		})

		When("invalid request", func() {

		})
	})

	Context("List", func() {
		When("valid request", func() {
			When("no resources", func() {
				It("returns empty list", func() {

				})
			})
		})
		When("invalid request", func() {

		})
	})

	Context("Put", func() {

	})

	Context("Delete", func() {

	})
})
