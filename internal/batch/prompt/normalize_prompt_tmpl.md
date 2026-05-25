제목: {{.Title}}

{{if .SaleInfo}}
판매처별 판매 정보
{{range .SaleInfo -}}
{{.Site}}
제목: {{.Title}}
{{- if .Desc}}
소개: {{.Desc}}
{{- end}}
{{- if .Series}}
상품과 같은 시리즈
{{- range .Series}}
- {{.}}
  {{- end}}
  {{- end}}

{{end}}
{{- end}}