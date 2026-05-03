package htmls

import (
	_ "embed"
	"strings"
)

const registerCodePlaceholder = "{{CODE}}"

//go:embed register-code-template-a1.html
var registerCodeTemplateA1 string

func RenderRegisterCodeTemplateA1(code string) string {
	return strings.ReplaceAll(registerCodeTemplateA1, registerCodePlaceholder, code)
}
