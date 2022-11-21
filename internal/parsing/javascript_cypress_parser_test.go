package parsing_test

import (
	"os"
	"strings"

	"github.com/rwx-research/captain-cli/internal/parsing"
	v1 "github.com/rwx-research/captain-cli/internal/testingschema/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JavaScriptCypressParser", func() {
	Describe("Parse", func() {
		It("parses the sample file", func() {
			fixture, err := os.Open("../../test/fixtures/cypress.xml")
			Expect(err).ToNot(HaveOccurred())

			parseResult, err := parsing.JavaScriptCypressParser{}.Parse(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(parseResult).NotTo(BeNil())

			Expect(parseResult.Parser).To(Equal(parsing.JavaScriptCypressParser{}))
			Expect(parseResult.Sentiment).To(Equal(parsing.PositiveParseResultSentiment))
			Expect(parseResult.TestResults.Framework.Language).To(Equal(v1.FrameworkLanguageJavaScript))
			Expect(parseResult.TestResults.Framework.Kind).To(Equal(v1.FrameworkKindCypress))
			Expect(parseResult.TestResults.Summary.Tests).To(Equal(0))
			Expect(parseResult.TestResults.Summary.OtherErrors).To(Equal(0))
		})

		It("errors on malformed XML", func() {
			parseResult, err := parsing.JavaScriptCypressParser{}.Parse(strings.NewReader(`<abc`))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unable to parse test results as XML"))
			Expect(parseResult).To(BeNil())
		})

		It("errors on XML that doesn't look like Cypress", func() {
			var parseResult *parsing.ParseResult
			var err error

			parseResult, err = parsing.JavaScriptCypressParser{}.Parse(strings.NewReader(`<foo></foo>`))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unable to parse test results as XML"))
			Expect(parseResult).To(BeNil())

			parseResult, err = parsing.JavaScriptCypressParser{}.Parse(strings.NewReader(`<testsuites></testsuites>`))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No tests count was found in the XML"))
			Expect(parseResult).To(BeNil())

			parseResult, err = parsing.JavaScriptCypressParser{}.Parse(
				strings.NewReader(`<testsuites tests="1"><testsuite></testsuite></testsuites>`),
			)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(
				ContainSubstring("The test suites in the XML do not appear to match Cypress XML"),
			)
			Expect(parseResult).To(BeNil())
		})
	})
})
