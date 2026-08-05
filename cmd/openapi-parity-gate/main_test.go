package main

import "testing"

func TestVerifyParityTestOutputRejectsSuccessfulPackageWithNoNamedTest(t *testing.T) {
	t.Parallel()

	output := []byte(`{"Action":"start","Package":"github.com/collectors-tech/cabinet/internal/app"}
{"Action":"pass","Package":"github.com/collectors-tech/cabinet/internal/app","Elapsed":0.1}
`)
	if err := verifyParityTestOutput(output); err == nil {
		t.Fatal("expected a successful go test package with no named parity test to be rejected")
	}
}

func TestVerifyParityTestOutputAcceptsNamedParitySuitePass(t *testing.T) {
	t.Parallel()

	output := []byte(`{"Action":"run","Package":"github.com/collectors-tech/cabinet/internal/app","Test":"TestOpenAPIParitySuite"}
{"Action":"pass","Package":"github.com/collectors-tech/cabinet/internal/app","Test":"TestOpenAPIParitySuite","Elapsed":0.1}
{"Action":"pass","Package":"github.com/collectors-tech/cabinet/internal/app","Elapsed":0.1}
`)
	if err := verifyParityTestOutput(output); err != nil {
		t.Fatalf("expected named parity suite pass to be accepted: %v", err)
	}
}
