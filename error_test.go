package errcode

import (
	"errors"
	"testing"
)

var (
	errFoo                  = errors.New("some error")
	gError                  = NewGeneralError("general error").Make()
	geDescendant1           = gError.Produce().SubType("descendant of general error").Make()
	geDescendant2           = gError.Produce().Message("geDescendant2").Make()
	geDescendant3           = geDescendant1.Produce().Message("geDescendant2").Make()
	geDescendantExtRegular  = gError.Produce().ExternalErrMess(errFoo).Make()
	geDescendant1ExtRegular = geDescendant1.Produce().ExternalErrMess(errFoo).Make()

	ee = errors.New("pepega is here")
	e  = NewGeneralError("Somebody").Make()
	e1 = e.Produce().SubType("Once").Make()
	e2 = e1.Produce().SubType("Told").Make()
	e3 = e2.Produce().SubType("Me").Make()
	e4 = e3.Produce().ExternalErrMess(ee).Make()
)

func Test_GeneralError(t *testing.T) {

	t.Run("gError.ErrorCode()", func(t *testing.T) {
		expectedCode := "1MWWTMD"
		if gError.ErrorCode() != expectedCode {
			t.Fatalf("gError.ErrorCode() got: %v, expected: %v", gError.ErrorCode(), expectedCode)
		}
	})

	t.Run("geDescendant1.ErrorCode()", func(t *testing.T) {
		expectedCode := "1IG3NBM"
		if geDescendant1.ErrorCode() != expectedCode {
			t.Fatalf("geDescendant1.ErrorCode() got: %v, expected: %v", geDescendant1.ErrorCode(), expectedCode)
		}
	})

	t.Run("geDescendant1.Is(gError)", func(t *testing.T) {
		if !geDescendant1.Is(gError) {
			t.Fatalf("geDescendant1.Is(gError) got: %v, expected: %v", false, true)
		}
	})

	t.Run("geDescendant2.Is(gError)", func(t *testing.T) {
		if !geDescendant2.Is(gError) {
			t.Fatalf("geDescendant2.Is(gError) got: %v, expected: %v", false, true)
		}
	})

	t.Run("geDescendant3.Is(geDescendant1)", func(t *testing.T) {
		if !geDescendant3.Is(geDescendant1) {
			t.Fatalf("geDescendant3.Is(geDescendant1) got: %v, expected: %v", false, true)
		}
	})

	t.Run("geDescendant3.Is(gError)", func(t *testing.T) {
		if !geDescendant3.Is(gError) {
			t.Fatalf("geDescendant3.Is(gError) got: %v, expected: %v", false, true)
		}
	})

	t.Run("geDescendantExtRegular.Is(gError)", func(t *testing.T) {
		if !geDescendantExtRegular.Is(gError) {
			t.Fatalf("geDescendantExtRegular.Is(gError) got: %v, expected: %v", false, true)
		}
	})

	t.Run("geDescendantExtRegular.Is(errFoo)", func(t *testing.T) {
		if !geDescendantExtRegular.Is(errFoo) {
			t.Fatalf("geDescendantExtRegular.Is(errFoo) got: %v, expected: %v", false, true)
		}
	})

	t.Run("geDescendant1ExtRegular.Is(gError)", func(t *testing.T) {
		if !geDescendant1ExtRegular.Is(gError) {
			t.Fatalf("geDescendant1ExtRegular.Is(gError) got: %v, expected: %v", false, true)
		}
	})

	t.Run("geDescendant1ExtRegular.Is(geDescendant1)", func(t *testing.T) {
		if !geDescendant1ExtRegular.Is(geDescendant1) {
			t.Fatalf("geDescendant1ExtRegular.Is(geDescendant1) got: %v, expected: %v", false, true)
		}
	})

	t.Run("geDescendant1ExtRegular.Is(errFoo)", func(t *testing.T) {
		if !geDescendant1ExtRegular.Is(errFoo) {
			t.Fatalf("geDescendant1ExtRegular.Is(errFoo) got: %v, expected: %v", false, true)
		}
	})

	t.Run("Descendants test", func(t *testing.T) {
		if e4.ErrorCode() != e3.ErrorCode() {
			t.Fatalf("e4.ErrorCode() != e3.ErrorCode() got: %v, expected: %v", false, true)
		}

		if e3.ErrorCode() == e2.ErrorCode() {
			t.Fatalf("e3.ErrorCode() == e2.ErrorCode() got: %v, expected: %v", true, false)
		}

		if e3.ErrorCode() == e1.ErrorCode() {
			t.Fatalf("e3.ErrorCode() == e1.ErrorCode() got: %v, expected: %v", true, false)
		}

		if e3.ErrorCode() == e.ErrorCode() {
			t.Fatalf("e3.ErrorCode() == e.ErrorCode() got: %v, expected: %v", true, false)
		}

		if e2.ErrorCode() == e1.ErrorCode() {
			t.Fatalf("e2.ErrorCode() == e1.ErrorCode() got: %v, expected: %v", true, false)
		}

		if e2.ErrorCode() == e.ErrorCode() {
			t.Fatalf("e2.ErrorCode() == e.ErrorCode() got: %v, expected: %v", true, false)
		}

		if e1.ErrorCode() == e.ErrorCode() {
			t.Fatalf("e1.ErrorCode() == e.ErrorCode() got: %v, expected: %v", true, false)
		}

		if !e4.Is(ee) {
			t.Fatalf("e4.Is(ee) got: %v, expected: %v", false, true)
		}

		if !e4.Is(e3) {
			t.Fatalf("e4.Is(e3) got: %v, expected: %v", false, true)
		}
		if !e4.Is(e2) {
			t.Fatalf("e4.Is(e2) got: %v, expected: %v", false, true)
		}

		if !e4.Is(e1) {
			t.Fatalf("e4.Is(e1) got: %v, expected: %v", false, true)
		}

		if !e4.Is(e) {
			t.Fatalf("e4.Is(e) got: %v, expected: %v", false, true)
		}
	})

}
