package docid

import "testing"

func has(got []string, want string) bool {
	for _, g := range got {
		if g == want {
			return true
		}
	}
	return false
}

func TestIdentifyContem(t *testing.T) {
	nfeBase := "3517081234567800019555001000000001100000001"
	nfe := nfeBase + string(rune('0'+nfeDV(nfeBase)))

	tests := []struct {
		in   string
		want string
	}{
		{"46706186015", "CPF"},
		{"11.222.333/0001-81", "CNPJ"},
		{"026.503.064-61", "CNH"},
		{"1234 5678 0191", "Título de Eleitor"},
		{"110.042.490.114", "Inscrição Estadual (SP)"},
		{"224/3658792", "Inscrição Estadual (RS)"},
		{"123.45678-50", "Inscrição Estadual (PR)"},
		{"ABC-1D23", "Placa Mercosul"},
		{"4006381333931", "EAN-13"},
		{"978-0-306-40615-7", "ISBN-13"},
		{"0-306-40615-2", "ISBN-10"},
		{"BR18 0036 0305 0000 1000 9795 493C 1", "IBAN"},
		{"GB82 WEST 1234 5698 7654 32", "IBAN"},
		{nfe, "Chave NF-e"},
		{"4111 1111 1111 1111", "credit-card-visa"},
		{"5555 5555 5555 4444", "credit-card-mastercard"},
		{"3782 822463 10005", "credit-card-amex"},
		{"6011 1111 1111 1117", "credit-card-discover"},
		{"3530 1113 3330 0000", "credit-card-jcb"},
		{"3056 930902 5904", "credit-card-diners"},
	}
	for _, tc := range tests {
		if got := Identify(tc.in); !has(got, tc.want) {
			t.Errorf("Identify(%q) = %v, esperava conter %q", tc.in, got, tc.want)
		}
	}
}

func TestIdentifyExato(t *testing.T) {
	got := Identify("467.061.860-15")
	if len(got) != 1 || got[0] != "CPF" {
		t.Errorf("esperava [CPF], veio %v", got)
	}
}

func TestIdentifyVazio(t *testing.T) {
	for _, in := range []string{"467.061.860-12", "11.222.333/0001-80", "ABC-1234", "", "abc"} {
		got := Identify(in)
		if got == nil || len(got) != 0 {
			t.Errorf("Identify(%q) = %#v, esperava slice vazio não nulo", in, got)
		}
	}
}
