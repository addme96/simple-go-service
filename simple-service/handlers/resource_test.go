package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	var (
		mockCtrl *gomock.Controller
		mockRepo *mocks.MockResourceRepository
		w        *httptest.ResponseRecorder
	)
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
				mockRepo.EXPECT().Create(req.Context(), r).Times(1).Return(nil)

				By("acting")
				handlers.NewResource(mockRepo).Post(w, req)

				By("asserting")
				res := w.Result()
				Expect(res.StatusCode).To(Equal(http.StatusCreated))
				defer res.Body.Close()
				resp, err := io.ReadAll(res.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp).To(BeEmpty())
			})

			When("repository errors", func() {
				It("returns 500", func() {
					By("arranging")
					r := entities.Resource{
						Name: "Resource Name",
					}
					body, err := json.Marshal(r)
					Expect(err).ShouldNot(HaveOccurred())
					req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
					req.Header.Set("Content-Type", "application/json")
					mockRepo.EXPECT().Create(req.Context(), r).Times(1).Return(errors.New("some err"))

					By("acting")
					handlers.NewResource(mockRepo).Post(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
					defer res.Body.Close()
					resp, err := io.ReadAll(res.Body)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(string(resp)).To(Equal("some err\n"))
				})
			})

			When("failed to read body", func() {
				It("returns 500", func() {
					By("arranging")
					req := httptest.NewRequest(http.MethodPost, "/", mocks.ErrReader{})
					req.Header.Set("Content-Type", "application/json")

					By("acting")
					handlers.NewResource(mockRepo).Post(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
					defer res.Body.Close()
					resp, err := io.ReadAll(res.Body)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(string(resp)).To(Equal("test error\n"))
				})
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

				By("asserting")
				res := w.Result()
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				defer res.Body.Close()
				resp, err := io.ReadAll(res.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp).To(BeEmpty())
			})
		})

		When("invalid body", func() {
			It("returns 400 Bad Request", func() {
				By("arranging")
				body := []byte("{{{ something invalid")
				req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

				By("acting")
				handlers.NewResource(mockRepo).Post(w, req)

				By("asserting")
				res := w.Result()
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				defer res.Body.Close()
				resp, err := io.ReadAll(res.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(string(resp)).To(Equal("invalid character '{' looking for beginning of object key string\n"))
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
				routeCtx := prepareRouteCtxWithURLParam("resourceID", strconv.Itoa(resourceID))
				req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(routeCtx)
				mockRepo.EXPECT().Read(req.Context(), resourceID).Times(1).Return(expectedResource, nil)

				By("acting")
				resourceHandler.GetCtx(nextHandler).ServeHTTP(w, req)

				By("asserting")
				res := w.Result()
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				Expect(resource).To(Equal(expectedResource))
			})
			It("returns 404 if resource not found", func() {
				By("arranging")
				resourceID := 123
				routeCtx := prepareRouteCtxWithURLParam("resourceID", strconv.Itoa(resourceID))
				req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(routeCtx)
				mockRepo.EXPECT().Read(req.Context(), resourceID).Times(1).Return(nil, pgx.ErrNoRows)

				By("acting")
				resourceHandler.GetCtx(nextHandler).ServeHTTP(w, req)

				By("asserting")
				res := w.Result()
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))
				Expect(resource).To(BeNil())
			})
		})
		When("invalid request", func() {
			It("returns 400 when not int", func() {
				By("arranging")
				routeCtx := prepareRouteCtxWithURLParam("resourceID", "definitely-not-int")
				req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(routeCtx)
				mockRepo.EXPECT().Read(gomock.Any(), gomock.Any()).Times(0)

				By("acting")
				resourceHandler.GetCtx(nextHandler).ServeHTTP(w, req)

				By("asserting")
				res := w.Result()
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

				By("asserting")
				res := w.Result()
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(body).To(MatchJSON(expectedBody))
			})

			It("returns 400 if there is no value in the context", func() {
				By("arranging")
				req := httptest.NewRequest(http.MethodGet, "/", nil)

				By("acting")
				handlers.NewResource(mockRepo).Get(w, req)

				By("asserting")
				res := w.Result()
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				defer res.Body.Close()
				resp, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(resp)).To(Equal("failed to read resource from the context\n"))
			})
		})

		When("invalid request", func() {
			It("returns 400 if invalid type in the context", func() {
				By("arranging")
				resourceID := 123
				resource := struct{ ID int }{resourceID}
				ctxWithResource := context.WithValue(context.TODO(), "resource", resource)
				req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctxWithResource)

				By("acting")
				handlers.NewResource(mockRepo).Get(w, req)

				By("asserting")
				res := w.Result()
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				defer res.Body.Close()
				resp, err := io.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(resp)).To(Equal("failed to read resource from the context\n"))
			})
		})
	})

	Context("List", func() {
		When("valid request", func() {
			When("no resources", func() {
				It("returns empty list", func() {
					By("arranging")
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					mockRepo.EXPECT().ReadAll(req.Context()).Times(1).Return([]entities.Resource{}, nil)

					By("acting")
					handlers.NewResource(mockRepo).List(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusOK))
					defer res.Body.Close()
					body, err := io.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(body).To(MatchJSON("[]"))
				})
			})

			When("single resource exists", func() {
				It("returns single resource", func() {
					By("arranging")
					resources := []entities.Resource{
						{
							ID:   123,
							Name: "Resource 1 Name",
						},
					}
					expectedBody, err := json.Marshal(resources)
					Expect(err).NotTo(HaveOccurred())
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					mockRepo.EXPECT().ReadAll(req.Context()).Times(1).Return(resources, nil)

					By("acting")
					handlers.NewResource(mockRepo).List(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusOK))
					defer res.Body.Close()
					body, err := io.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(body).To(MatchJSON(expectedBody))
				})
			})

			When("two resources exist", func() {
				It("returns both resources", func() {
					By("arranging")
					resources := []entities.Resource{
						{
							ID:   123,
							Name: "Resource 1 Name",
						},
						{
							ID:   456,
							Name: "Resource 2 Name",
						},
					}
					expectedBody, err := json.Marshal(resources)
					Expect(err).NotTo(HaveOccurred())
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					mockRepo.EXPECT().ReadAll(req.Context()).Times(1).Return(resources, nil)

					By("acting")
					handlers.NewResource(mockRepo).List(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusOK))
					defer res.Body.Close()
					body, err := io.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(body).To(MatchJSON(expectedBody))
				})
			})
			When("repository errors", func() {
				It("returns 500", func() {
					By("arranging")
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					mockRepo.EXPECT().ReadAll(req.Context()).Times(1).Return(nil, fmt.Errorf("error"))

					By("acting")
					handlers.NewResource(mockRepo).List(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
					defer res.Body.Close()
					resp, err := io.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resp)).To(Equal("error\n"))
				})
			})
		})
	})

	Context("Put", func() {
		When("valid request", func() {
			It("updates the resource", func() {
				By("arranging")
				currentResource := entities.Resource{
					ID:   123,
					Name: "Resource Name",
				}
				ctxWithResource := context.WithValue(context.TODO(), "resource", &currentResource)
				newResource := entities.Resource{
					Name: "Resource Name Changed",
				}
				body, err := json.Marshal(newResource)
				Expect(err).ShouldNot(HaveOccurred())
				req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body)).WithContext(ctxWithResource)
				req.Header.Set("Content-Type", "application/json")
				mockRepo.EXPECT().Update(req.Context(), currentResource.ID, newResource).Times(1).Return(nil)

				By("acting")
				handlers.NewResource(mockRepo).Put(w, req)

				By("asserting")
				res := w.Result()
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				defer res.Body.Close()
				resp, err := io.ReadAll(res.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp).To(BeEmpty())
			})

			When("failed to read body", func() {
				It("returns 500", func() {
					By("arranging")
					currentResource := entities.Resource{
						ID:   123,
						Name: "Resource Name",
					}
					ctxWithResource := context.WithValue(context.TODO(), "resource", &currentResource)
					req := httptest.NewRequest(http.MethodPut, "/", mocks.ErrReader{}).WithContext(ctxWithResource)
					req.Header.Set("Content-Type", "application/json")

					By("acting")
					handlers.NewResource(mockRepo).Put(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
					defer res.Body.Close()
					resp, err := io.ReadAll(res.Body)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(string(resp)).To(Equal("test error\n"))
				})
			})
		})

		When("invalid Content-Type", func() {
			It("returns 400 Bad Request", func() {
				By("arranging")
				currentResource := entities.Resource{
					ID:   123,
					Name: "Resource Name",
				}
				ctxWithResource := context.WithValue(context.TODO(), "resource", &currentResource)
				newResource := entities.Resource{
					Name: "Resource Name Changed",
				}
				body, err := json.Marshal(newResource)
				Expect(err).ShouldNot(HaveOccurred())
				req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body)).WithContext(ctxWithResource)
				mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
				Expect(err).ShouldNot(HaveOccurred())

				By("acting")
				handlers.NewResource(mockRepo).Put(w, req)

				By("asserting")
				res := w.Result()
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				defer res.Body.Close()
				resp, err := io.ReadAll(res.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(string(resp)).To(Equal("invalid Content-Type - should be application/json\n"))
			})

			When("repository errors", func() {
				It("returns 500", func() {
					By("arranging")
					currentResource := entities.Resource{
						ID:   123,
						Name: "Resource Name",
					}
					ctxWithResource := context.WithValue(context.TODO(), "resource", &currentResource)
					newResource := entities.Resource{
						Name: "Resource Name Changed",
					}
					body, err := json.Marshal(newResource)
					req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(body)).WithContext(ctxWithResource)
					req.Header.Set("Content-Type", "application/json")
					mockRepo.EXPECT().Update(req.Context(), currentResource.ID, newResource).Times(1).Return(errors.New("some err"))

					By("acting")
					handlers.NewResource(mockRepo).Put(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
					defer res.Body.Close()
					resp, err := io.ReadAll(res.Body)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(string(resp)).To(Equal("some err\n"))
				})
			})
		})

		When("invalid request", func() {
			When("invalid body", func() {
				It("returns 400 Bad Request", func() {
					By("arranging")
					currentResource := entities.Resource{
						ID:   123,
						Name: "Resource Name",
					}
					ctxWithResource := context.WithValue(context.TODO(), "resource", &currentResource)
					body := []byte("{{{ something invalid")
					req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body)).WithContext(ctxWithResource)
					req.Header.Set("Content-Type", "application/json")
					mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

					By("acting")
					handlers.NewResource(mockRepo).Put(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
					defer res.Body.Close()
					resp, err := io.ReadAll(res.Body)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(string(resp)).To(Equal("invalid character '{' looking for beginning of object key string\n"))
				})
			})

			When("invalid type in the context", func() {
				It("returns 400", func() {
					By("arranging")
					resourceID := 123
					resource := struct{ ID int }{resourceID}
					ctxWithResource := context.WithValue(context.TODO(), "resource", resource)
					req := httptest.NewRequest(http.MethodPut, "/", nil).WithContext(ctxWithResource)
					req.Header.Set("Content-Type", "application/json")
					By("acting")
					handlers.NewResource(mockRepo).Put(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
					defer res.Body.Close()
					resp, err := io.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resp)).To(Equal("failed to read resource from the context\n"))
				})
			})
		})
	})

	Context("Delete", func() {
		When("valid request", func() {
			It("deletes the resource", func() {
				By("arranging")
				resource := entities.Resource{
					ID:   123,
					Name: "Resource Name",
				}
				ctxWithResource := context.WithValue(context.TODO(), "resource", &resource)
				req := httptest.NewRequest(http.MethodDelete, "/", nil).WithContext(ctxWithResource)
				req.Header.Set("Content-Type", "application/json")
				mockRepo.EXPECT().Delete(req.Context(), resource.ID).Times(1).Return(nil)

				By("acting")
				handlers.NewResource(mockRepo).Delete(w, req)

				By("asserting")
				res := w.Result()
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				defer res.Body.Close()
				resp, err := io.ReadAll(res.Body)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(resp).To(BeEmpty())
			})

			When("repository errors", func() {
				It("returns 500", func() {
					By("arranging")
					resource := entities.Resource{
						Name: "Resource Name",
					}
					ctxWithResource := context.WithValue(context.TODO(), "resource", &resource)
					req := httptest.NewRequest(http.MethodDelete, "/", nil).WithContext(ctxWithResource)
					mockRepo.EXPECT().Delete(req.Context(), resource.ID).Times(1).Return(errors.New("some err"))

					By("acting")
					handlers.NewResource(mockRepo).Delete(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
					defer res.Body.Close()
					resp, err := io.ReadAll(res.Body)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(string(resp)).To(Equal("some err\n"))
				})
			})
		})

		When("invalid request", func() {
			When("invalid type in the context", func() {
				It("returns 400", func() {
					By("arranging")
					resourceID := 123
					resource := struct{ ID int }{resourceID}
					ctxWithResource := context.WithValue(context.TODO(), "resource", resource)
					req := httptest.NewRequest(http.MethodDelete, "/", nil).WithContext(ctxWithResource)
					req.Header.Set("Content-Type", "application/json")
					By("acting")
					handlers.NewResource(mockRepo).Delete(w, req)

					By("asserting")
					res := w.Result()
					Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
					defer res.Body.Close()
					resp, err := io.ReadAll(res.Body)
					Expect(err).NotTo(HaveOccurred())
					Expect(string(resp)).To(Equal("failed to read resource from the context\n"))
				})
			})
		})
	})
})

func prepareRouteCtxWithURLParam(key, val string) context.Context {
	routeParams := chi.RouteParams{}
	routeParams.Add(key, val)
	routeContext := chi.Context{URLParams: routeParams}
	return context.WithValue(context.TODO(), chi.RouteCtxKey, &routeContext)
}
