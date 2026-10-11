# docid

Package em Go que identifica o **tipo de um documento ou código** (CPF, CNPJ, CNH, NF-e, IBAN, cartões etc.) a partir do valor informado, com ou sem pontuação. A identificação valida o formato e os dígitos verificadores de cada padrão.

```go
docid.Identify("467.061.860-15") // ["CPF"]
docid.Identify("467.061.860-12") // []
```

## Instalação

```bash
go get github.com/antonioscairato/docid
```

## Uso

```go
package main

import (
	"fmt"

	"github.com/antonioscairato/docid"
)

func main() {
	fmt.Println(docid.Identify("11.222.333/0001-81"))     // [CNPJ]
	fmt.Println(docid.Identify("4111 1111 1111 1111"))    // [credit-card-visa]
	fmt.Println(docid.Identify("978-0-306-40615-7"))      // [ISBN-13 ...]
	fmt.Println(docid.Identify("valor inválido"))         // []
}
```

### Como funciona

`func Identify(code string) []string`

- Pontuação, espaços e separadores (`.`, `-`, `/`, espaços etc.) são ignorados, e maiúsculas/minúsculas não importam.
- O retorno é um **slice com todos os padrões compatíveis**, na ordem de prioridade.
- Se nenhum padrão for compatível, retorna um slice **vazio** (nunca `nil`).

### Ambiguidades

Alguns códigos satisfazem mais de um padrão por coincidência dos dígitos verificadores. Por exemplo, um número de 11 dígitos pode ser, ao mesmo tempo, CPF, PIS, CNH e/ou RENAVAM. Por isso o retorno é uma lista: use outras informações do seu contexto para decidir qual interpretação é a correta.

```go
docid.Identify("026.503.064-61") // ["PIS/PASEP/NIT", "CNH"]
```

## Padrões suportados

| Rótulo retornado | Padrão | Observações |
|---|---|---|
| `CPF` | CPF | 11 dígitos, 2 DVs |
| `CNPJ` | CNPJ | 14 caracteres; suporta o formato numérico e o **alfanumérico** |
| `CNH` | Carteira Nacional de Habilitação | 11 dígitos, 2 DVs |
| `Título de Eleitor` | Título de Eleitor | 12 dígitos; valida UF (01–28) e DVs |
| `PIS/PASEP/NIT` | PIS / PASEP / NIT | 11 dígitos, módulo 11 |
| `RENAVAM` | RENAVAM | 11 dígitos (ou 9, formato antigo) |
| `Chave NF-e` | Chave de acesso da NF-e | 44 dígitos; valida código da UF e DV |
| `Inscrição Estadual (SP)`, `(PR)`, `(RS)` | Inscrição Estadual | Apenas SP, PR e RS (veja abaixo) |
| `Placa Mercosul` | Placa de veículo | Formato `LLLNLNN` (não há dígito verificador) |
| `credit-card-<bandeira>` | Cartões de crédito/débito | Luhn + faixa e tamanho; veja abaixo |
| `EAN-13`, `EAN-8`, `UPC-A` | Código de barras | Dígito verificador GS1 |
| `ISBN-13`, `ISBN-10` | ISBN | ISBN-13 exige prefixo 978 ou 979 |
| `IBAN` | IBAN | Módulo 97; valida o tamanho de países comuns |

### Bandeiras de cartão

O retorno é `credit-card-` seguido da bandeira: `visa`, `mastercard`, `amex`, `diners`, `discover`, `jcb`, `elo`, `hipercard`, `maestro` ou `unionpay`.

- Cartões com Luhn válido, mas de bandeira não reconhecida, **não** são retornados (para reduzir falsos positivos).
- O rótulo sempre começa com `credit-card-`, mesmo para cartões de débito, pois o número não distingue os dois.
- As faixas de Elo, Hipercard e Maestro usam uma lista simplificada de prefixos.

## Limitações

- **Inscrição Estadual:** cada estado tem sua própria regra. Hoje só SP, PR e RS estão implementados. Para adicionar outro estado, inclua uma função em `ieValidators` e a sigla em `ieOrder` (arquivo `docid.go`).
- **RG:** não é suportado, pois o formato varia entre os estados e não há algoritmo único. Com a Carteira de Identidade Nacional (CIN), o número passa a ser o CPF, que já é reconhecido.
- **Validação formal, não existência:** o package confirma que o código é estruturalmente válido, não que ele esteja de fato emitido, ativo ou pertença a alguém.
- **Dados sensíveis:** CPF, CNH, cartões e outros códigos são dados pessoais. Evite registrá-los em logs.

## Testes

```bash
go test ./...
```

## Licença

Distribuído sob a licença MIT. Veja o arquivo [LICENSE](LICENSE).
