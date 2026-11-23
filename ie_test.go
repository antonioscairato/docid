package docid

import "testing"

// Exemplos de IE válidas (SINTEGRA / bibliotecas de validação), uma por UF/formato.
var ieVectors = []struct{ uf, ie string }{
	{"AC", "01.004.823/001-12"},
	{"AL", "240000048"},
	{"AP", "030123459"},
	{"BA", "123456-63"},
	{"BA", "612345-57"},
	{"BA", "1000003-06"},
	{"CE", "06000001-5"},
	{"DF", "07300001001-09"},
	{"ES", "99999999-0"},
	{"GO", "10.987.654-7"},
	{"MA", "12000038-5"},
	{"MT", "0013000001-9"},
	{"MG", "062.307.904/0081"},
	{"PA", "15-999999-5"},
	{"PB", "06000001-5"},
	{"PR", "123.45678-50"},
	{"PE", "0321418-40"},
	{"PE", "18.1.001.0000004-9"},
	{"PI", "01234567-9"},
	{"RJ", "99.999.99-3"},
	{"RN", "20.040.040-1"},
	{"RN", "20.0.040.040-0"},
	{"RS", "224/3658792"},
	{"RO", "0000000062521-3"},
	{"RO", "101.62521-3"},
	{"RR", "24006628-1"},
	{"SC", "251.040.852"},
	{"SP", "110.042.490.114"},
	{"SE", "27123456-3"},
	{"TO", "29.01.022783-6"},
}

func TestIEPorUF(t *testing.T) {
	for _, v := range ieVectors {
		want := "Inscrição Estadual (" + v.uf + ")"
		if got := Identify(v.ie); !has(got, want) {
			t.Errorf("%s %q: esperava %q em %v", v.uf, v.ie, want, got)
		}
	}
}

// Alterar o último dígito (DV) de uma IE válida deve invalidá-la para aquela UF.
func TestIEDVAdulterado(t *testing.T) {
	for _, v := range ieVectors {
		n := []byte(normalize(v.ie))
		n[len(n)-1] = '0' + (n[len(n)-1]-'0'+1)%10
		want := "Inscrição Estadual (" + v.uf + ")"
		if got := Identify(string(n)); has(got, want) {
			t.Errorf("%s %q com DV adulterado (%s) ainda foi aceita", v.uf, v.ie, n)
		}
	}
}

// BA: o DV1 (penúltimo dígito) também precisa ser conferido.
func TestIEBAPenultimoDV(t *testing.T) {
	for _, ie := range []string{"123456-63", "612345-57", "1000003-06"} {
		n := []byte(normalize(ie))
		i := len(n) - 2
		n[i] = '0' + (n[i]-'0'+1)%10
		if validIEBA(string(n)) {
			t.Errorf("BA %q com DV1 adulterado (%s) ainda foi aceita", ie, n)
		}
	}
}

func TestIEPrefixos(t *testing.T) {
	// Exigem prefixo fixo: mesmo com DV correto, outro prefixo é recusado.
	if validIEAC("0200482300112") {
		t.Error("AC deve exigir prefixo 01")
	}
	if validIEAL("250000048") {
		t.Error("AL deve exigir prefixo 24")
	}
}
