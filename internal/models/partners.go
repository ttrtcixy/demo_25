package models

type Partner struct {
	Id          int
	CompanyName string
	PartnerType string
	Director    string
	Phone       string
	Rating      int
	Sale        int

	Email   string
	Address string

	Discount int
}

type Partners []Partner
