package grouter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gavrilaf/grouter"
)

var _ = Describe("grouter", func() {
	var subject Router

	Describe("Add route", func() {
		var err error

		BeforeEach(func() {
			subject = NewRouter()
		})

		It("should add simple url", func() {
			err = subject.AddRoute("GET", "https://api.github.com/search/repositories", 1)
			Expect(err).To(BeNil())
		})

		It("should return an error when adding two identical URLs", func() {
			err = subject.AddRoute("GET", "https://api.github.com/search/repositories", 1)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/search/repositories", 1)
			Expect(err).To(Equal(ErrAlreadyAdded))
		})

		It("should add parameterized url", func() {
			err = subject.AddRoute("GET", "https://api.github.com/applications/grants/:grant_id", 1)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/users/:username/events", 2)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/users/vasya/events", 2)
			Expect(err).To(BeNil())
		})

		It("should add catch all url", func() {
			err = subject.AddRoute("GET", "https://aadhi.cma.r53.nordstrom.net:443/v1/authtoken/*", 1)
			Expect(err).To(BeNil())
		})

		It("should add parameterized catch all url", func() {
			err = subject.AddRoute("GET", "https://api.github.com/v1/authtoken/*some", 1)
			Expect(err).To(BeNil())
		})

		It("should add url with parameterized query", func() {
			err = subject.AddRoute("GET", "https://api.github.com/v1/authtoken?user=:user_id&api_key=*&format=json", 1)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/v1/authtoken?user=:user_id&api_key=*&format=xml", 2)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/repos/*?format=json&token=*&id=:id", 3)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/repos/*?format=json&token=*", 4)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/repos/*?token=*&format=xml", 5)
			Expect(err).To(BeNil())
		})

		It("should return error when two different varibles on same place", func() {
			err = subject.AddRoute("GET", "https://api.github.com/applications/grants/:grant_id/no", 1)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/applications/grants/:other_id/no", 1)
			Expect(err).ToNot(BeNil())
		})

		It("should return error when variable conflicts with catchAll", func() {
			err = subject.AddRoute("GET", "https://api.github.com/applications/grants/:grant_id/no", 1)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/applications/grants/*", 1)
			Expect(err).ToNot(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/applications/events/*", 1)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/applications/events/:event_id", 1)
			Expect(err).ToNot(BeNil())
		})

		It("should return error when catchAll variable conflicts with catchAll", func() {
			err = subject.AddRoute("GET", "https://api.github.com/applications/grants/*path", 1)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/applications/grants/*", 1)
			Expect(err).ToNot(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/applications/events/*", 1)
			Expect(err).To(BeNil())

			err = subject.AddRoute("GET", "https://api.github.com/applications/events/*path", 1)
			Expect(err).ToNot(BeNil())
		})
	})

	Describe("Lookup", func() {
		var item *ParsedRoute

		BeforeEach(func() {
			subject = NewRouter()
		})

		Describe("Root", func() {
			BeforeEach(func() {
				subject.AddRoute("GET", "https://api.github.com", 10)
			})

			It("should match root url", func() {
				item, _ = subject.Lookup("GET", "https://api.github.com/")
				Expect(item.Value).To(Equal(10))
			})
		})

		Describe("Equality", func() {
			BeforeEach(func() {
				subject.AddRoute("GET", "https://api.github.com/search/repositories", 1)
			})

			It("should match url by equality", func() {
				item, _ = subject.Lookup("get", "https://api.github.com/search/repositories")
				Expect(item.Value).To(Equal(1))
			})

			It("should match url with http schema (http & https are the same for router)", func() {
				item, _ = subject.Lookup("get", "http://api.github.com/search/repositories")
				Expect(item.Value).To(Equal(1))
			})

			It("should not match url with unknown host", func() {
				item, _ = subject.Lookup("GET", "https://facebook.com/search/repositories")
				Expect(item).To(BeNil())
			})

			It("should not match url with unknown endpoint", func() {
				item, _ = subject.Lookup("GET", "https://api.github.com/update/repositories")
				Expect(item).To(BeNil())
			})

			It("should not match url with unknown method", func() {
				item, _ = subject.Lookup("POST", "https://api.github.com/search/repositories")
				Expect(item).To(BeNil())
			})
		})

		Describe("Url wildcard", func() {
			BeforeEach(func() {
				subject.AddRoute("POST", "https://api.github.com/users/:username/events", 2)
				subject.AddRoute("POST", "https://api.github.com/users/vasya/events", 3)
			})

			It("should match url with wildcard", func() {
				item, _ = subject.Lookup("post", "https://api.github.com/users/john-doe/events")
				Expect(item.Value).To(Equal(2))

				expected := map[string]string{"username": "john-doe"}
				Expect(item.UrlParams).To(Equal(expected))
			})

			It("Direct url has priority on wildcard", func() {
				item, _ = subject.Lookup("POST", "https://api.github.com/users/vasya/events")
				Expect(item.Value).To(Equal(3))
				Expect(item.UrlParams).To(BeEmpty())
			})
		})

		Describe("Catch all wildcard", func() {
			BeforeEach(func() {
				subject.AddRoute("PUT", "https://api.github.com/authorizations/clients/*client", 4)
				subject.AddRoute("GET", "https://api.github.com/authorizations/events/*", 5)
			})

			// https://api.github.com/authorizations/events/*
			It("should match catch all wildcard", func() {
				item, _ = subject.Lookup("GET", "https://api.github.com/authorizations/events/1")
				Expect(item.Value).To(Equal(5))

				item, _ = subject.Lookup("GET", "https://api.github.com/authorizations/events/1/2/3")
				Expect(item.Value).To(Equal(5))
			})

			// https://api.github.com/authorizations/clients/*client
			It("should match named catch all wildcard", func() {
				item, _ = subject.Lookup("PUT", "https://api.github.com/authorizations/clients/client-1")
				expected := map[string]string{"client": "client-1"}

				Expect(item.Value).To(Equal(4))
				Expect(item.UrlParams).To(Equal(expected))

				item, _ = subject.Lookup("PUT", "https://api.github.com/authorizations/clients/client-22/fingerprint")
				expected = map[string]string{"client": "client-22/fingerprint"}

				Expect(item.Value).To(Equal(4))
				Expect(item.UrlParams).To(Equal(expected))
			})
		})

		Describe("Query string", func() {
			BeforeEach(func() {
				subject.AddRoute("GET", "https://api.github.com/repos/*?format=json&token=*&id=:id", 6)
				subject.AddRoute("GET", "https://api.github.com/repos/*?format=json&token=:token", 7)
				subject.AddRoute("GET", "https://api.github.com/repos/*?token=*&format=xml", 8)
			})

			It("should match url with query", func() {
				// https://api.github.com/repos/*?format=json&token=*&id=:id
				item, _ = subject.Lookup("GET", "https://api.github.com/repos/repo-1?format=json&token=123456&id=12")
				expected := map[string]string{"id": "12"}

				Expect(item.Value).To(Equal(6))
				Expect(item.QueryParams).To(Equal(expected))

				// https://api.github.com/repos/*?format=json&token=:token
				item, _ = subject.Lookup("GET", "https://api.github.com/repos/repo-1/update?format=json&token=8797")
				expected = map[string]string{"token": "8797"}

				Expect(item.Value).To(Equal(7))
				Expect(item.QueryParams).To(Equal(expected))

				// https://api.github.com/repos/*?token=*&format=xml
				item, _ = subject.Lookup("GET", "https://api.github.com/repos/repo-2?format=xml&token=1234")

				Expect(item.Value).To(Equal(8))
				Expect(item.QueryParams).To(BeEmpty())

			})

			It("should not match url with query", func() {
				// https://api.github.com/repos/*?format=json&token=*&id=:id; format=json & id=:id & token=* or format=xml & token=* (no id)
				item, _ = subject.Lookup("GET", "https://api.github.com/repos/repo-2?format=xml&token=1234&id=78")
				Expect(item).To(BeNil())

			})
		})

		Describe("Query string, catch all", func() {
			BeforeEach(func() {
				subject.AddRoute("GET", "https://test.net:443/secure.google.com/v1/authinit?format=json&*", 1)
			})

			It("should match url with query", func() {
				item, _ = subject.Lookup("GET", "https://test.net:443/secure.google.com/v1/authinit?format=json&token=12")
				Expect(item.Value).To(Equal(1))

				item, _ = subject.Lookup("GET", "https://test.net:443/secure.google.com/v1/authinit?format=json&token=12&code=9876")
				Expect(item.Value).To(Equal(1))
			})

			It("should not match url with query", func() {
				item, _ = subject.Lookup("GET", "https://test.net:443/secure.google.com/v1/authinit?format=xml&token=12")
				Expect(item).To(BeNil())

			})
		})

		Describe("Query string, realword examples", func() {
			BeforeEach(func() {
				// The following urls has difference only in query params
				subject.AddRoute("GET", "https://test.net/disco/breadcrumb/offers?orderby=Boosted&breadcrumb=Home/Men/All%20Men&category=mens-view-all", 101)
				subject.AddRoute("GET", "https://test.net/disco/breadcrumb/offers?orderby=Boosted&breadcrumb=Home/Men/All%20Men&category=mens-view-all&filterby=store%20eq%201", 102)
				subject.AddRoute("GET", "https://test.net/disco/breadcrumb/offers?orderby=Boosted&breadcrumb=Home/Men/All%20Men&category=mens-view-all&filterby=searchcolorfacet%20eq%20'Black'", 103)
				subject.AddRoute("GET", "https://test.net/disco/breadcrumb/offers?orderby=Boosted&breadcrumb=Home/Men/All%20Men&category=mens-view-all&filterby=searchcolorfacet%20eq%20'Black'%5Estore%20eq%201", 104)

				// This url crashed router - fixed
				subject.AddRoute("GET", "https://test.net/v1.2/styleservice/style/4618153/shippingdescription?format=json&apikey=*&postalcode=", 105)
			})

			It("should match url with query", func() {
				item, _ = subject.Lookup("GET", "https://test.net/disco/breadcrumb/offers?orderby=Boosted&breadcrumb=Home/Men/All%20Men&category=mens-view-all")
				Expect(item.Value).To(Equal(101))

				item, _ = subject.Lookup("GET", "https://test.net/disco/breadcrumb/offers?orderby=Boosted&breadcrumb=Home/Men/All%20Men&category=mens-view-all&filterby=store%20eq%201")
				Expect(item.Value).To(Equal(102))

				item, _ = subject.Lookup("GET", "https://test.net/disco/breadcrumb/offers?orderby=Boosted&breadcrumb=Home/Men/All%20Men&category=mens-view-all&filterby=searchcolorfacet%20eq%20'Black'")
				Expect(item.Value).To(Equal(103))

				item, _ = subject.Lookup("GET", "https://test.net/disco/breadcrumb/offers?orderby=Boosted&breadcrumb=Home/Men/All%20Men&category=mens-view-all&filterby=searchcolorfacet%20eq%20'Black'%5Estore%20eq%201")
				Expect(item.Value).To(Equal(104))

				item, _ = subject.Lookup("GET", "https://test.net/v1.2/styleservice/style/4618153/shippingdescription?format=json&apikey=GQZExhNLtY7e4kiFCuZAaw72rkSUcFuY&postalcode=")
				Expect(item.Value).To(Equal(105))
			})
		})
	})
})
