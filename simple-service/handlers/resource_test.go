package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/addme96/simple-go-service/simple-service/entities"
	"github.com/addme96/simple-go-service/simple-service/handlers"
	"github.com/addme96/simple-go-service/simple-service/handlers/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource", func() {
	var mockCtrl *gomock.Controller
	var mockRepo *mocks.MockResourceRepository
	var w *httptest.ResponseRecorder
	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockRepo = mocks.NewMockResourceRepository(mockCtrl)
		w = httptest.NewRecorder()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("NewResource", func() {
		It("creates resource handler with a given repository", func() {
			handler := handlers.NewResource(mockRepo)
			Expect(*handler).To(Equal(handlers.Resource{Repository: mockRepo}))
		})
	})

	Context("Post", func() {
		When("valid request", func() {
			It("creates the resource", func() {
				By("arranging")
				r := entities.Resource{
					Name: "Resource Name",
				}
				body, err := json.Marshal(r)
				Expect(err).ShouldNot(HaveOccurred())
				req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				mockRepo.EXPECT().Create(gomock.Any(), r).Times(1).Return(nil)

				By("acting")
				handlers.NewResource(mockRepo).Post(w, req)
				res := w.Result()
				defer res.Body.Close()
				resp, err := io.ReadAll(res.Body)

				By("asserting")
				Expect(res.StatusCode).To(Equal(http.StatusCreated))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp).To(BeEmpty())
			})
		})

		When("invalid Content-Type", func() {
			It("returns 400 Bad Request", func() {
				By("arranging")

				r := entities.Resource{
					Name: "Resource Name",
				}
				body, err := json.Marshal(r)
				Expect(err).ShouldNot(HaveOccurred())
				req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)
				Expect(err).ShouldNot(HaveOccurred())

				By("acting")
				handlers.NewResource(mockRepo).Post(w, req)
				res := w.Result()
				defer res.Body.Close()
				resp, err := io.ReadAll(res.Body)

				By("asserting")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp).To(BeEmpty())
			})
		})
	})

	Context("GetCtx", func() {
		var resource *entities.Resource
		var resourceHandler *handlers.Resource
		var nextHandler http.HandlerFunc
		BeforeEach(func() {
			resourceHandler = handlers.NewResource(mockRepo)
			nextHandler = func(writer http.ResponseWriter, request *http.Request) {
				resource = request.Context().Value("resource").(*entities.Resource)
			}
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
				routeCtx := prepareRouteCtxWithURLParam("resourceID", strconv.Itoa(resourceID))
				req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(routeCtx)

				By("acting")
				resourceHandler.GetCtx(nextHandler).ServeHTTP(w, req)
				res := w.Result()

				By("asserting")
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(resource).To(Equal(expectedResource))
			})
			It("returns 404 if resource not found", func() {
				By("arranging")
				resourceID := 123
				mockRepo.EXPECT().Read(gomock.Any(), resourceID).Times(1).Return(nil, pgx.ErrNoRows)
				routeCtx := prepareRouteCtxWithURLParam("resourceID", strconv.Itoa(resourceID))
				req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(routeCtx)

				By("acting")
				resourceHandler.GetCtx(nextHandler).ServeHTTP(w, req)
				res := w.Result()

				By("asserting")
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				Expect(resource).To(BeNil())
			})
		})
		When("invalid request", func() {
			It("returns 400 when not int", func() {
				By("arranging")
				mockRepo.EXPECT().Read(gomock.Any(), gomock.Any()).Times(0)
				routeCtx := prepareRouteCtxWithURLParam("resourceID", "definitely-not-int")
				req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(routeCtx)

				By("acting")
				resourceHandler.GetCtx(nextHandler).ServeHTTP(w, req)
				res := w.Result()

				By("asserting")
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
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
				req := httptest.NewRequest(http.MethodGet, "/", nil)

				By("acting")
				handlers.NewResource(mockRepo).Get(w, req)
				res := w.Result()
				defer res.Body.Close()
				resp, err := io.ReadAll(res.Body)
				By("asserting")
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).To(BeEmpty())
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

func prepareRouteCtxWithURLParam(key, val string) context.Context {
	routeParams := chi.RouteParams{}
	routeParams.Add(key, val)
	routeContext := chi.Context{URLParams: routeParams}
	return context.WithValue(context.TODO(), chi.RouteCtxKey, &routeContext)
}
