package models

type Mensagem struct {
	Source    string `json:"source"`
	Barcode   string `json:"barcode"`
	DataEnvio string `json:"data_envio"`
	Code      string `json:"code"`
}
