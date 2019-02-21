package grouter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gavrilaf/grouter"
)

var _ = Describe("grouter", func() {
	var subject Router

	BeforeEach(func() {
		subject = NewRouter()
	})

	Describe("Add route", func() {
		var err error

		// Add test for the root url (only host, without path)

		It("Should add simple url", func() {
			err = subject.AddRoute("https://api.github.com/search/repositories", 1)
			Expect(err).To(BeNil())
		})

		It("Should add parameterized url", func() {
			err = subject.AddRoute("https://api.github.com/applications/grants/:grant_id", 1)
			Expect(err).To(BeNil())

			err = subject.AddRoute("https://api.github.com/users/:username/events", 2)
			Expect(err).To(BeNil())

			err = subject.AddRoute("https://api.github.com/users/vasya/events", 2)
			Expect(err).To(BeNil())
		})

		It("Should add catch all url", func() {
			err = subject.AddRoute("https://aadhi.cma.r53.nordstrom.net:443/v1/authtoken/*", 1)
			Expect(err).To(BeNil())
		})

		It("Should add parameterized catch all url", func() {
			err = subject.AddRoute("https://api.github.com/v1/authtoken/*some", 1)
			Expect(err).To(BeNil())
		})

		It("Should add url with parameterized query", func() {
			err = subject.AddRoute("https://api.github.com/v1/authtoken?user=:user_id&api_key=*&format=json", 1)
			Expect(err).To(BeNil())

			err = subject.AddRoute("https://api.github.com/v1/authtoken?user=:user_id&api_key=*&format=xml", 2)
			Expect(err).To(BeNil())

			err = subject.AddRoute("https://api.github.com/repos/*?format=json&token=*&id=:id", 3)
			Expect(err).To(BeNil())

			err = subject.AddRoute("https://api.github.com/repos/*?format=json&token=*", 4)
			Expect(err).To(BeNil())

			err = subject.AddRoute("https://api.github.com/repos/*?token=*&format=xml", 5)
			Expect(err).To(BeNil())
		})

		// Add errors test
	})

	Describe("Lookup", func() {
		//var err error
		var item *ParsedRoute

		BeforeEach(func() {
			subject.AddRoute("https://api.github.com/search/repositories", 1)

			subject.AddRoute("https://api.github.com/users/:username/events", 2)
			subject.AddRoute("https://api.github.com/users/vasya/events", 3)

			subject.AddRoute("https://api.github.com/authorizations/clients/*client", 4)
			subject.AddRoute("https://api.github.com/authorizations/events/*", 5)

			subject.AddRoute("https://api.github.com/repos/*?format=json&token=*&id=:id", 6)
			subject.AddRoute("https://api.github.com/repos/*?format=json&token=*", 7)
			subject.AddRoute("https://api.github.com/repos/*?token=*&format=xml", 8)
		})

		It("Should find url by equality", func() {
			item, _ = subject.Lookup("https://api.github.com/search/repositories")

			Expect(item.Value).To(Equal(1))
		})

		It("Should find parameterized url", func() {
			item, _ = subject.Lookup("https://api.github.com/users/john-doe/events")

			Expect(item.Value).To(Equal(2))

			expected := map[string]string{"username": "john-doe"}
			Expect(item.UrlParams).To(Equal(expected))
		})

		It("Direct url has priority on parameterized", func() {
			item, _ = subject.Lookup("https://api.github.com/users/vasya/events")

			Expect(item.Value).To(Equal(3))
			Expect(item.UrlParams).To(BeEmpty())
		})

		It("Should find parameterized catch all url", func() {
			item, _ = subject.Lookup("https://api.github.com/authorizations/clients/client-1")

			Expect(item.Value).To(Equal(4))

			expected := map[string]string{"client": "client-1"}
			Expect(item.UrlParams).To(Equal(expected))

			item, _ = subject.Lookup("https://api.github.com/authorizations/clients/client-22/fingerprint")

			Expect(item.Value).To(Equal(4))

			expected = map[string]string{"client": "client-22/fingerprint"}
			Expect(item.UrlParams).To(Equal(expected))
		})

		It("Should find catch all url", func() {
			item, _ = subject.Lookup("https://api.github.com/authorizations/events/1")
			Expect(item.Value).To(Equal(5))

			item, _ = subject.Lookup("https://api.github.com/authorizations/events/1/2/3")
			Expect(item.Value).To(Equal(5))
		})

		It("Should find url with query", func() {
			item, _ = subject.Lookup("https://api.github.com/repos/repo-1?format=json&token=123456&id=12")
			Expect(item.Value).To(Equal(6))

			item, _ = subject.Lookup("https://api.github.com/repos/repo-1/update?format=json&token=8797")
			Expect(item.Value).To(Equal(7))

			item, _ = subject.Lookup("https://api.github.com/repos/repo-2?format=xml&token=1234")
			Expect(item.Value).To(Equal(8))

			item, _ = subject.Lookup("https://api.github.com/repos/repo-2?format=xml&token=1234&xid=78")
			Expect(item.Value).To(Equal(8))
		})
	})
})