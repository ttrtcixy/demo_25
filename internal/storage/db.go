package storage

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ttrtcixy/demo/internal/models"
	"log"
	"math"
	"time"
)

type DB struct {
	connect *sql.DB
}
type Query struct {
	query string
	args  []any
}

func NewDB() *DB {
	d, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatalln(err)
	}

	return &DB{connect: d}
}

// var getPartners = `select PartnerId, PartnerType, PartnerName, Director, Phone, Rating, Email, LegalAddress From Partners;`
var getPartners = `SELECT 
   p.PartnerId, p.PartnerType, p.PartnerName, p.Director, p.Phone, p.Rating, p.Email, p.LegalAddress,
    CASE
        WHEN SUM(pp.Quantity) < 10000 THEN 0
        WHEN SUM(pp.Quantity) BETWEEN 10000 AND 49999 THEN 5
        WHEN SUM(pp.Quantity) BETWEEN 50000 AND 299999 THEN 10
        WHEN SUM(pp.Quantity) >= 300000 THEN 15
    END AS DiscountPercentage
FROM 
    Partners p
JOIN 
    PartnerProducts pp ON p.PartnerId = pp.PartnerId
GROUP BY 
    p.PartnerId, p.PartnerName;`

var ErrPartnersNoFound = errors.New("партнеры не найдены")

func (d *DB) GetPartners() (*models.Partners, error) {
	query := Query{query: getPartners}
	rows, err := d.connect.Query(query.query)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return &models.Partners{}, ErrPartnersNoFound
	}

	partners := models.Partners{}
	for {
		var partner models.Partner
		err := rows.Scan(&partner.Id, &partner.PartnerType, &partner.CompanyName, &partner.Director, &partner.Phone, &partner.Rating, &partner.Email, &partner.Address, &partner.Discount)
		if err != nil {
			return nil, err
		}
		partners = append(partners, partner)

		if !rows.Next() {
			break
		}
	}

	return &partners, nil
}

var addPartner = `insert into Partners(PartnerType, PartnerName, Director, Phone, Rating, Email, LegalAddress) values(?, ?, ?, ?, ?, ?, ?)`

func (d *DB) AddPartner(partner models.Partner) error {
	args := []any{partner.PartnerType, partner.CompanyName, partner.Director, partner.Phone, partner.Rating, partner.Email, partner.Address}
	query := Query{query: addPartner, args: args}
	_, err := d.connect.Exec(query.query, query.args...)
	if err != nil {
		return err
	}
	return nil
}

var deletePartner = `delete from Partners where PartnerId = ?`

func (d *DB) DeletePartner(id int) error {
	query := Query{query: deletePartner, args: []any{id}}
	_, err := d.connect.Exec(query.query, query.args...)
	if err != nil {
		return err
	}
	return nil
}

var updatePartner = `update Partners set PartnerType = ?, PartnerName = ?, Director = ?, Phone = ?, Rating = ? where PartnerId = ?;`

func (d *DB) UpdatePartner(partner models.Partner) error {
	args := []any{partner.PartnerType, partner.CompanyName, partner.Director, partner.Phone, partner.Rating, partner.Id}
	query := Query{query: updatePartner, args: args}
	_, err := d.connect.Exec(query.query, query.args...)
	if err != nil {
		return err
	}
	return nil
}

var getPartnerSales = `
    SELECT 
        p.ProductName AS 'Продукция',
        pp.Quantity AS 'Количество',
        pp.SaleDate AS 'Дата продажи',
        pt.ProductType AS 'Тип продукции',
        (pp.Quantity * p.MinCost) AS 'Общая сумма'
    FROM 
        PartnerProducts pp
    JOIN 
        Products p ON pp.ProductId = p.ProductId
    JOIN 
        ProductTypes pt ON p.ProductTypeId = pt.ProductTypeId
    WHERE 
        pp.PartnerId = ?
    ORDER BY 
        pp.SaleDate DESC`

func (d *DB) GetPartnerSales(id int) ([]models.PartnerSale, error) {
	query := `
    SELECT 
        p.ProductName AS 'Продукция',
        pp.Quantity AS 'Количество',
        pp.SaleDate AS 'Дата продажи',
        pt.ProductType AS 'Тип продукции',
        (pp.Quantity * p.MinCost) AS 'Общая сумма'
    FROM 
        PartnerProducts pp
    JOIN 
        Products p ON pp.ProductId = p.ProductId
    JOIN 
        ProductTypes pt ON p.ProductTypeId = pt.ProductTypeId
    WHERE 
        pp.PartnerId = ?
    ORDER BY 
        pp.SaleDate DESC`

	var sales []models.PartnerSale
	rows, err := d.connect.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса продаж: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sale models.PartnerSale
		var rawDate interface{} // Принимаем дату как интерфейс

		err := rows.Scan(
			&sale.ProductName,
			&sale.Quantity,
			&rawDate,
			&sale.ProductType,
			&sale.TotalSum,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки: %v", err)
		}

		// Преобразуем дату в строку
		switch v := rawDate.(type) {
		case time.Time:
			sale.SaleDate = v.Format("2006-01-02")
		case []byte:
			sale.SaleDate = string(v)
		case string:
			sale.SaleDate = v
		default:
			sale.SaleDate = fmt.Sprintf("%v", v)
		}
		sales = append(sales, sale)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при обработке результатов: %v", err)
	}

	return sales, nil
}

func (d *DB) FindPartner(searchTerm string) (int, string, error) {
	var partnerID int
	var partnerName string

	// Пробуем найти по ID
	err := d.connect.QueryRow("SELECT PartnerId, PartnerName FROM Partners WHERE PartnerId = ?", searchTerm).Scan(&partnerID, &partnerName)
	if err == nil {
		return partnerID, partnerName, nil
	}

	// Если не нашли по ID, ищем по имени
	err = d.connect.QueryRow("SELECT PartnerId, PartnerName FROM Partners WHERE PartnerName LIKE ? LIMIT 1", "%"+searchTerm+"%").Scan(&partnerID, &partnerName)
	if err != nil {
		return 0, "", fmt.Errorf("партнер не найден")
	}

	return partnerID, partnerName, nil
}

func (db *DB) GetProducts() ([]string, error) {
	rows, err := db.connect.Query("SELECT ProductId, ProductName FROM Products")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []string
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		products = append(products, fmt.Sprintf("%s - %s", id, name))
	}
	return products, nil
}

func (db *DB) GetMaterialTypes() ([]string, error) {
	rows, err := db.connect.Query("SELECT MaterialTypeId, MaterialType FROM MaterialTypes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var materials []string
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		materials = append(materials, fmt.Sprintf("%s - %s", id, name))
	}
	return materials, nil
}

func (db *DB) CalculateMaterial(productId, materialId string, quantity int, param1, param2 float64) (int, error) {
	// Получаем коэффициент продукта
	var productCoef float64
	err := db.connect.QueryRow(`
        SELECT pt.Coefficient 
        FROM ProductTypes pt
        JOIN Products p ON pt.ProductTypeId = p.ProductTypeId
        WHERE p.ProductId = ?`, productId).Scan(&productCoef)
	if err != nil {
		return -1, fmt.Errorf("не найден коэффициент для продукта")
	}

	// Получаем процент брака
	var defectPercentage float64
	err = db.connect.QueryRow(`
        SELECT DefectPercentage 
        FROM MaterialTypes 
        WHERE MaterialTypeId = ?`, materialId).Scan(&defectPercentage)
	if err != nil {
		return -1, fmt.Errorf("не найден процент брака для материала")
	}

	// Расчет
	materialPerUnit := param1 * param2 * productCoef
	totalMaterial := float64(quantity) * materialPerUnit
	if defectPercentage > 0 {
		totalMaterial = totalMaterial * (1 + defectPercentage/100)
	}

	return int(math.Ceil(totalMaterial)), nil
}
