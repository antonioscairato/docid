// Package docid identifica o tipo de um documento/código (CPF, CNPJ, CNH,
// NF-e, RENAVAM, IBAN, ISBN, EAN etc.) a partir do valor informado, com ou
// sem pontuação, validando o formato e os dígitos verificadores.
package docid

import "strings"

// validator devolve os rótulos compatíveis com o código (nil se nenhum).
type validator func(s string) []string

func named(name string, f func(string) bool) validator {
	return func(s string) []string {
		if f(s) {
			return []string{name}
		}
		return nil
	}
}

// Ordem de prioridade dos resultados: padrões mais específicos primeiro.
var validators = []validator{
	named("Chave NF-e", validNFe),
	named("IBAN", validIBAN),
	named("Placa Mercosul", validPlacaMercosul),
	named("CNPJ", validCNPJ),
	named("CPF", validCPF),
	named("ISBN-13", validISBN13),
	named("ISBN-10", validISBN10),
	named("EAN-13", func(s string) bool { return len(s) == 13 && validGS1(s) }),
	named("UPC-A", func(s string) bool { return len(s) == 12 && validGS1(s) }),
	named("EAN-8", func(s string) bool { return len(s) == 8 && validGS1(s) }),
	named("Título de Eleitor", validTituloEleitor),
	named("PIS/PASEP/NIT", validPIS),
	named("CNH", validCNH),
	named("RENAVAM", validRENAVAM),
	ieLabels,
	cardBrands,
}

// Identify retorna todos os tipos compatíveis com o código informado
// (pontuação e separadores são ignorados). Se nada for compatível, retorna
// um slice vazio. Há ambiguidades reais: um número de 11 dígitos pode ser
// CPF, PIS, CNH e/ou RENAVAM ao mesmo tempo, por coincidência dos DVs.
func Identify(code string) []string {
	s := normalize(code)
	out := []string{}
	for _, v := range validators {
		out = append(out, v(s)...)
	}
	return out
}

// ---------------------------------------------------------------- helpers

// normalize remove tudo que não for letra ou dígito ASCII e põe em maiúsculas.
func normalize(code string) string {
	var b strings.Builder
	for _, r := range code {
		switch {
		case r >= '0' && r <= '9', r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= 'a' && r <= 'z':
			b.WriteRune(r - 'a' + 'A')
		}
	}
	return b.String()
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func allEqual(s string) bool {
	for i := 1; i < len(s); i++ {
		if s[i] != s[0] {
			return false
		}
	}
	return true
}

func digits(s string) []int {
	d := make([]int, len(s))
	for i := 0; i < len(s); i++ {
		d[i] = int(s[i] - '0')
	}
	return d
}

// dv11 é o módulo 11 "clássico": resto < 2 -> 0, senão 11 - resto.
func dv11(sum int) int {
	r := sum % 11
	if r < 2 {
		return 0
	}
	return 11 - r
}

func weightedSum(d, w []int) int {
	sum := 0
	for i := range w {
		sum += d[i] * w[i]
	}
	return sum
}

// ---------------------------------------------------------------- CPF

func validCPF(s string) bool {
	if len(s) != 11 || !isDigits(s) || allEqual(s) {
		return false
	}
	d := digits(s)
	return d[9] == cpfDV(d[:9], 10) && d[10] == cpfDV(d[:10], 11)
}

func cpfDV(d []int, start int) int {
	sum := 0
	for i, v := range d {
		sum += v * (start - i)
	}
	r := (sum * 10) % 11
	if r == 10 {
		return 0
	}
	return r
}

// ---------------------------------------------------------------- CNPJ
// Suporta o CNPJ numérico e o alfanumérico (a partir de julho/2026): os 12
// primeiros caracteres podem ser letras ou dígitos (valor = código ASCII - 48)
// e os 2 últimos são sempre dígitos verificadores.

var (
	cnpjW1 = []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	cnpjW2 = []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
)

func validCNPJ(s string) bool {
	if len(s) != 14 || allEqual(s) || !isDigits(s[12:]) {
		return false
	}
	v := make([]int, 14)
	for i := 0; i < 14; i++ {
		v[i] = int(s[i] - '0') // 'A' -> 17, 'B' -> 18, ...
	}
	return v[12] == dv11(weightedSum(v, cnpjW1)) &&
		v[13] == dv11(weightedSum(v, cnpjW2))
}

// ---------------------------------------------------------------- CNH

func validCNH(s string) bool {
	if len(s) != 11 || !isDigits(s) || allEqual(s) {
		return false
	}
	d := digits(s)

	sum := 0
	for i := 0; i < 9; i++ {
		sum += d[i] * (9 - i)
	}
	dv1, dsc := sum%11, 0
	if dv1 >= 10 {
		dv1, dsc = 0, 2
	}

	sum = 0
	for i := 0; i < 9; i++ {
		sum += d[i] * (i + 1)
	}
	dv2 := sum%11 - dsc
	if dv2 < 0 {
		dv2 += 11
	}
	if dv2 >= 10 {
		dv2 = 0
	}
	return d[9] == dv1 && d[10] == dv2
}

// ---------------------------------------------------------------- Título de Eleitor
// 8 dígitos sequenciais + 2 dígitos de UF (01..28) + 2 dígitos verificadores.

func validTituloEleitor(s string) bool {
	if len(s) != 12 || !isDigits(s) {
		return false
	}
	d := digits(s)
	uf := d[8]*10 + d[9]
	if uf < 1 || uf > 28 {
		return false
	}
	special := uf == 1 || uf == 2 // SP e MG

	fix := func(sum int) int {
		r := sum % 11
		if r == 10 {
			return 0
		}
		if r == 0 && special {
			return 1
		}
		return r
	}

	sum := 0
	for i := 0; i < 8; i++ {
		sum += d[i] * (i + 2)
	}
	dv1 := fix(sum)
	dv2 := fix(d[8]*7 + d[9]*8 + dv1*9)
	return d[10] == dv1 && d[11] == dv2
}

// ---------------------------------------------------------------- PIS/PASEP/NIT

var pisW = []int{3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

func validPIS(s string) bool {
	if len(s) != 11 || !isDigits(s) || allEqual(s) {
		return false
	}
	d := digits(s)
	return d[10] == dv11(weightedSum(d, pisW))
}

// ---------------------------------------------------------------- RENAVAM
// Aceita 11 dígitos (ou 9, formato antigo, completado com zeros à esquerda).

var renavamW = []int{3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

func validRENAVAM(s string) bool {
	if !isDigits(s) || (len(s) != 11 && len(s) != 9) || allEqual(s) {
		return false
	}
	s = strings.Repeat("0", 11-len(s)) + s
	d := digits(s)
	dv := (weightedSum(d, renavamW) * 10) % 11
	if dv == 10 {
		dv = 0
	}
	return d[10] == dv
}

// ---------------------------------------------------------------- NF-e

var ufCodes = map[string]bool{
	"11": true, "12": true, "13": true, "14": true, "15": true, "16": true, "17": true,
	"21": true, "22": true, "23": true, "24": true, "25": true, "26": true, "27": true,
	"28": true, "29": true, "31": true, "32": true, "33": true, "35": true,
	"41": true, "42": true, "43": true, "50": true, "51": true, "52": true, "53": true,
}

func nfeDV(first43 string) int {
	sum, w := 0, 2
	for i := len(first43) - 1; i >= 0; i-- {
		sum += int(first43[i]-'0') * w
		if w++; w > 9 {
			w = 2
		}
	}
	return dv11(sum)
}

func validNFe(s string) bool {
	if len(s) != 44 || !isDigits(s) || !ufCodes[s[:2]] {
		return false
	}
	return int(s[43]-'0') == nfeDV(s[:43])
}

// ---------------------------------------------------------------- Placa Mercosul (LLLNLNN)
// Não há dígito verificador: só o formato.

func isLetter(c byte) bool { return c >= 'A' && c <= 'Z' }
func isDigit(c byte) bool  { return c >= '0' && c <= '9' }

func validPlacaMercosul(s string) bool {
	return len(s) == 7 &&
		isLetter(s[0]) && isLetter(s[1]) && isLetter(s[2]) &&
		isDigit(s[3]) && isLetter(s[4]) && isDigit(s[5]) && isDigit(s[6])
}

// ---------------------------------------------------------------- Cartões (Luhn)

func validLuhn(s string) bool {
	if len(s) < 13 || len(s) > 19 || !isDigits(s) || allEqual(s) {
		return false
	}
	sum, double := 0, false
	for i := len(s) - 1; i >= 0; i-- {
		n := int(s[i] - '0')
		if double {
			if n *= 2; n > 9 {
				n -= 9
			}
		}
		sum += n
		double = !double
	}
	return sum%10 == 0
}

// ---------------------------------------------------------------- Bandeira do cartão
// Luhn válido + faixa (prefixo) e tamanho conhecidos. Retorna, por exemplo,
// "credit-card-visa". Cartões com Luhn válido mas bandeira desconhecida não
// são considerados (para reduzir falsos positivos).

func prefixInt(s string, n int) int {
	v := 0
	for i := 0; i < n; i++ {
		v = v*10 + int(s[i]-'0')
	}
	return v
}

func hasPrefix(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func lenIn(s string, lengths ...int) bool {
	for _, l := range lengths {
		if len(s) == l {
			return true
		}
	}
	return false
}

func between(v, lo, hi int) bool { return v >= lo && v <= hi }

type cardRule struct {
	brand string
	match func(s string) bool
}

// Ordem importa: bandeiras com faixas "dentro" de outras vêm antes
// (Elo/Hipercard antes de Visa/Discover/Mastercard/UnionPay).
var cardRules = []cardRule{
	{"elo", func(s string) bool {
		return len(s) == 16 && hasPrefix(s, "401178", "401179", "438935", "457631", "457632",
			"431274", "451416", "457393", "504175", "5067", "509", "627780", "636297",
			"636368", "650", "6516", "6550")
	}},
	{"hipercard", func(s string) bool {
		return lenIn(s, 16, 19) && hasPrefix(s, "606282", "384100", "384140", "384160")
	}},
	{"amex", func(s string) bool { return len(s) == 15 && hasPrefix(s, "34", "37") }},
	{"diners", func(s string) bool {
		return between(len(s), 14, 19) &&
			(between(prefixInt(s, 3), 300, 305) || hasPrefix(s, "36", "38", "39"))
	}},
	{"discover", func(s string) bool {
		return between(len(s), 16, 19) &&
			(hasPrefix(s, "6011", "65") || between(prefixInt(s, 3), 644, 649) ||
				between(prefixInt(s, 6), 622126, 622925))
	}},
	{"jcb", func(s string) bool {
		return between(len(s), 16, 19) && between(prefixInt(s, 4), 3528, 3589)
	}},
	{"mastercard", func(s string) bool {
		return len(s) == 16 &&
			(between(prefixInt(s, 2), 51, 55) || between(prefixInt(s, 4), 2221, 2720))
	}},
	{"visa", func(s string) bool { return lenIn(s, 13, 16, 19) && hasPrefix(s, "4") }},
	{"maestro", func(s string) bool {
		return hasPrefix(s, "5018", "5020", "5038", "5893", "6304", "6759", "6761", "6762", "6763")
	}},
	{"unionpay", func(s string) bool { return between(len(s), 16, 19) && hasPrefix(s, "62") }},
}

func cardBrands(s string) []string {
	if !validLuhn(s) {
		return nil
	}
	for _, r := range cardRules {
		if r.match(s) {
			return []string{"credit-card-" + r.brand}
		}
	}
	return nil
}

// ---------------------------------------------------------------- EAN-8 / UPC-A / EAN-13 (GS1)

func validGS1(s string) bool {
	if !isDigits(s) {
		return false
	}
	n, sum := len(s), 0
	for i := n - 2; i >= 0; i-- {
		d := int(s[i] - '0')
		if (n-2-i)%2 == 0 {
			sum += d * 3
		} else {
			sum += d
		}
	}
	return (10-sum%10)%10 == int(s[n-1]-'0')
}

// ---------------------------------------------------------------- ISBN

func validISBN13(s string) bool {
	return len(s) == 13 && (strings.HasPrefix(s, "978") || strings.HasPrefix(s, "979")) && validGS1(s)
}

func validISBN10(s string) bool {
	if len(s) != 10 || !isDigits(s[:9]) {
		return false
	}
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(s[i]-'0') * (10 - i)
	}
	switch {
	case s[9] == 'X':
		sum += 10
	case isDigit(s[9]):
		sum += int(s[9] - '0')
	default:
		return false
	}
	return sum%11 == 0
}

// ---------------------------------------------------------------- IBAN

// Tamanhos oficiais de alguns países; países fora da tabela aceitam 15..34.
var ibanLen = map[string]int{
	"BR": 29, "PT": 25, "ES": 24, "FR": 27, "DE": 22, "IT": 27, "GB": 22,
	"NL": 18, "BE": 16, "CH": 21, "IE": 22, "LU": 20, "AT": 20, "PL": 28,
	"SE": 24, "NO": 15, "DK": 18, "FI": 18, "AR": 28, "TR": 26, "SA": 24,
}

func validIBAN(s string) bool {
	if len(s) < 15 || len(s) > 34 {
		return false
	}
	if !isLetter(s[0]) || !isLetter(s[1]) || !isDigit(s[2]) || !isDigit(s[3]) {
		return false
	}
	if n, ok := ibanLen[s[:2]]; ok && n != len(s) {
		return false
	}
	rem := 0
	for _, c := range s[4:] + s[:4] {
		switch {
		case c >= '0' && c <= '9':
			rem = (rem*10 + int(c-'0')) % 97
		case c >= 'A' && c <= 'Z':
			rem = (rem*100 + int(c-'A') + 10) % 97
		default:
			return false
		}
	}
	return rem == 1
}

// ---------------------------------------------------------------- Inscrição Estadual
// Cada UF tem sua regra. Aqui estão SP, PR e RS; para os demais estados,
// basta adicionar uma função em ieValidators.

var ieValidators = map[string]func(string) bool{
	"SP": validIESP,
	"PR": validIEPR,
	"RS": validIERS,
}

// ieOrder fixa a ordem de verificação (o mapa não tem ordem garantida).
var ieOrder = []string{"SP", "PR", "RS"}

func ieLabels(s string) []string {
	var out []string
	for _, uf := range ieOrder {
		if ieValidators[uf](s) {
			out = append(out, "Inscrição Estadual ("+uf+")")
		}
	}
	return out
}

// SP: 12 dígitos; DV1 na posição 9 e DV2 na posição 12 (módulo 11, %10).
func validIESP(s string) bool {
	if len(s) != 12 || !isDigits(s) {
		return false
	}
	d := digits(s)
	w1 := []int{1, 3, 4, 5, 6, 7, 8, 10}
	w2 := []int{3, 2, 10, 9, 8, 7, 6, 5, 4, 3, 2}
	return d[8] == weightedSum(d, w1)%11%10 && d[11] == weightedSum(d, w2)%11%10
}

// PR: 10 dígitos (8 + 2 DVs, módulo 11).
func validIEPR(s string) bool {
	if len(s) != 10 || !isDigits(s) {
		return false
	}
	d := digits(s)
	w1 := []int{3, 2, 7, 6, 5, 4, 3, 2}
	w2 := []int{4, 3, 2, 7, 6, 5, 4, 3, 2}
	return d[8] == dv11(weightedSum(d, w1)) && d[9] == dv11(weightedSum(d, w2))
}

// RS: 10 dígitos (9 + 1 DV).
func validIERS(s string) bool {
	if len(s) != 10 || !isDigits(s) {
		return false
	}
	d := digits(s)
	w := []int{2, 9, 8, 7, 6, 5, 4, 3, 2}
	dv := 11 - weightedSum(d, w)%11
	if dv >= 10 {
		dv = 0
	}
	return d[9] == dv
}
