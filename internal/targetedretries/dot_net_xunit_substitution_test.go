package targetedretries_test

import (
	"os"
	"sort"

	"github.com/bradleyjkemp/cupaloy"

	"github.com/rwx-research/captain-cli/internal/parsing"
	"github.com/rwx-research/captain-cli/internal/targetedretries"
	v1 "github.com/rwx-research/captain-cli/internal/testingschema/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DotNetxUnitSubstitution", func() {
	It("adheres to the Substitution interface", func() {
		var substitution targetedretries.Substitution = targetedretries.DotNetxUnitSubstitution{}
		Expect(substitution).NotTo(BeNil())
	})

	It("works with a real file", func() {
		substitution := targetedretries.DotNetxUnitSubstitution{}
		compiledTemplate, compileErr := targetedretries.CompileTemplate(substitution.Example())
		Expect(compileErr).NotTo(HaveOccurred())

		err := substitution.ValidateTemplate(compiledTemplate)
		Expect(err).NotTo(HaveOccurred())

		fixture, err := os.Open("../../test/fixtures/xunit_dot_net.xml")
		Expect(err).ToNot(HaveOccurred())

		testResults, err := parsing.DotNetxUnitParser{}.Parse(fixture)
		Expect(err).ToNot(HaveOccurred())

		substitutions := substitution.SubstitutionsFor(compiledTemplate, *testResults)
		sort.SliceStable(substitutions, func(i int, j int) bool {
			return substitutions[i]["filter"] < substitutions[j]["filter"]
		})
		cupaloy.SnapshotT(GinkgoT(), substitutions)
	})

	Describe("Example", func() {
		It("compiles and is valid", func() {
			substitution := targetedretries.DotNetxUnitSubstitution{}
			compiledTemplate, compileErr := targetedretries.CompileTemplate(substitution.Example())
			Expect(compileErr).NotTo(HaveOccurred())

			err := substitution.ValidateTemplate(compiledTemplate)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("ValidateTemplate", func() {
		It("is invalid for a template without placeholders", func() {
			substitution := targetedretries.DotNetxUnitSubstitution{}
			compiledTemplate, compileErr := targetedretries.CompileTemplate("dotnet test --filter")
			Expect(compileErr).NotTo(HaveOccurred())

			err := substitution.ValidateTemplate(compiledTemplate)
			Expect(err).To(HaveOccurred())
		})

		It("is invalid for a template with too many placeholders", func() {
			substitution := targetedretries.DotNetxUnitSubstitution{}
			compiledTemplate, compileErr := targetedretries.CompileTemplate("dotnet test --filter '{{ filter }}' {{ other }}")
			Expect(compileErr).NotTo(HaveOccurred())

			err := substitution.ValidateTemplate(compiledTemplate)
			Expect(err).To(HaveOccurred())
		})

		It("is invalid for a template without a filter placeholder", func() {
			substitution := targetedretries.DotNetxUnitSubstitution{}
			compiledTemplate, compileErr := targetedretries.CompileTemplate("dotnet test --filter '{{ other }}'")
			Expect(compileErr).NotTo(HaveOccurred())

			err := substitution.ValidateTemplate(compiledTemplate)
			Expect(err).To(HaveOccurred())
		})

		It("is valid for a template with only a filter placeholder", func() {
			substitution := targetedretries.DotNetxUnitSubstitution{}
			compiledTemplate, compileErr := targetedretries.CompileTemplate("dotnet test --filter '{{ filter }}'")
			Expect(compileErr).NotTo(HaveOccurred())

			err := substitution.ValidateTemplate(compiledTemplate)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Substitutions", func() {
		It("returns the unique test type.method", func() {
			compiledTemplate, compileErr := targetedretries.CompileTemplate("dotnet test --filter '{{ filter }}'")
			Expect(compileErr).NotTo(HaveOccurred())

			type1 := "type1"
			method1 := "method1"
			type2 := "type2"
			method2 := "method2"
			type3 := "type3"
			method3 := "method3"
			type4 := "type4"
			method4 := "method4"
			type5 := "type5"
			method5 := "method5"
			type6 := "type6"
			method6 := "method6"
			testResults := v1.TestResults{
				Tests: []v1.Test{
					{
						Attempt: v1.TestAttempt{
							Meta:   map[string]any{"type": &type1, "method": &method1},
							Status: v1.NewFailedTestStatus(nil, nil, nil),
						},
					},
					{
						Attempt: v1.TestAttempt{
							Meta:   map[string]any{"type": &type2, "method": &method2},
							Status: v1.NewCanceledTestStatus(),
						},
					},
					{
						Attempt: v1.TestAttempt{
							Meta:   map[string]any{"type": &type3, "method": &method3},
							Status: v1.NewTimedOutTestStatus(),
						},
					},
					{
						Attempt: v1.TestAttempt{
							Meta:   map[string]any{"type": &type3, "method": &method3},
							Status: v1.NewFailedTestStatus(nil, nil, nil),
						},
					},
					{
						Attempt: v1.TestAttempt{
							Meta:   map[string]any{"type": &type4, "method": &method4},
							Status: v1.NewPendedTestStatus(nil),
						},
					},
					{
						Attempt: v1.TestAttempt{
							Meta:   map[string]any{"type": &type5, "method": &method5},
							Status: v1.NewSuccessfulTestStatus(),
						},
					},
					{
						Attempt: v1.TestAttempt{
							Meta:   map[string]any{"type": &type6, "method": &method6},
							Status: v1.NewSkippedTestStatus(nil),
						},
					},
				},
			}

			substitution := targetedretries.DotNetxUnitSubstitution{}
			Expect(substitution.SubstitutionsFor(compiledTemplate, testResults)).To(Equal(
				[]map[string]string{
					{
						"filter": "FullyQualifiedName=type1.method1 | " +
							"FullyQualifiedName=type2.method2 | FullyQualifiedName=type3.method3",
					},
				},
			))
		})
	})
})
