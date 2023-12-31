package productsRepositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/NatthawutSK/ri-shop/config"
	"github.com/NatthawutSK/ri-shop/modules/entities"
	"github.com/NatthawutSK/ri-shop/modules/files/filesUsecases"
	"github.com/NatthawutSK/ri-shop/modules/products"
	"github.com/NatthawutSK/ri-shop/modules/products/productsPatterns"
	"github.com/jmoiron/sqlx"
)

type IProductsRepository interface{
	FindOneProduct(productId string) (*products.Products, error)
	FindProduct(req *products.ProductFilter) ([]*products.Products, int)
	InsertProduct(req *products.Products) (*products.Products, error)
	UpdateProduct(req *products.Products) (*products.Products, error)
	DeleteProduct(productId string) error
}

type productsRepository struct {
	db *sqlx.DB
	cfg config.IConfig
	fileUsecase filesUsecases.IFilesUsecase
}

func ProductsRepository(db *sqlx.DB, cfg config.IConfig, fileUsecase filesUsecases.IFilesUsecase) IProductsRepository {
	return &productsRepository{
		db: db,
		cfg: cfg,
		fileUsecase: fileUsecase,
	}
}

func (r *productsRepository) FindOneProduct(productId string) (*products.Products, error) {
	query := `
	SELECT
		to_jsonb("t")
	FROM (
		SELECT
			"p"."id",
			"p"."title",
			"p"."description",
			"p"."price",
			(
				SELECT
					to_jsonb("ct")
				FROM (
					SELECT
						"c"."id",
						"c"."title"
					FROM "categories" "c"
						LEFT JOIN "products_categories" "pc" ON "pc"."category_id" = "c"."id"
					WHERE "pc"."product_id" = "p"."id"
				) AS "ct"
			) AS "category",
			"p"."created_at",
			"p"."updated_at",
			(
				SELECT
					COALESCE(array_to_json(array_agg("it")), '[]'::json)
				FROM (
					SELECT
						"i"."id",
						"i"."filename",
						"i"."url"
					FROM "images" "i"
					WHERE "i"."product_id" = "p"."id"
				) AS "it"
			) AS "images"
		FROM "products" "p"
		WHERE "p"."id" = $1
		LIMIT 1
	) AS "t";`

	//COALESCE(array_to_json(array_agg("it")), '[]'::json) 
	// คือ ถ้าไม่มีข้อมูล(null) ให้ return '[]'::json แทน

	// array_agg คือ การรวมข้อมูลใน array ที่มีค่าเหมือนกันเป็น 1 row
	// array_to_json คือ การแปลง array เป็น json
	// to_jsonb คือ การแปลง NON-JSON เป็น jsonb


	productBytes := make([]byte, 0)
	product := &products.Products{
		Images: make([]*entities.Image, 0), //เวลาสร้าง struct ใหม่ แล้วข้างในมี array ให้ make array ไว้เลยเพื่อป้องกัน null pointer
	}
	if err := r.db.Get(&productBytes, query, productId); err != nil {
		return nil, fmt.Errorf("get product failed: %v", err)
	}
	if err := json.Unmarshal(productBytes, &product); err != nil {
		return nil, fmt.Errorf("unmarshal product failed: %v", err)
	}



	return product, nil  

}


func (r *productsRepository) FindProduct(req *products.ProductFilter) ([]*products.Products, int) {
	builder := productsPatterns.FindProductBuilder(r.db, req)
	engineer := productsPatterns.FindProductEngineer(builder)

	result := engineer.FindProduct().Result()
	count := engineer.CountProduct().Count()

	engineer.FindProduct().PrintQuery()

	return result, count
}


func (r *productsRepository) InsertProduct(req *products.Products) (*products.Products, error) {
	builder := productsPatterns.InsertProductBuilder(r.db, req)
	productId, err := productsPatterns.InsertProductEngineer(builder).InsertProduct()
	if err != nil {
		return nil, fmt.Errorf("insert product failed: %v", err)
	}

	product, err := r.FindOneProduct(productId)
	if err != nil {
		return nil, fmt.Errorf("find product failed: %v", err)
	}

	return product, nil
	
}

func (r *productsRepository) UpdateProduct(req *products.Products) (*products.Products, error) {
	builder := productsPatterns.UpdateProductBuilder(r.db, req, r.fileUsecase, r.cfg)
	engineer := productsPatterns.UpdateProductEngineer(builder)
	
	if err := engineer.UpdateProduct(); err != nil {
		return nil, err
	}

	product, err := r.FindOneProduct(req.Id)
	if err != nil {
		return nil,  err
	}
	return product, nil

}

func (r *productsRepository) DeleteProduct(productId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 15) // Timeout of 15 seconds
	defer cancel()
	query := `DELETE FROM "products" WHERE "id" = $1;`

	if _, err := r.db.ExecContext(ctx, query, productId); err != nil {
    	return fmt.Errorf("delete product failed: %v", err)
	}

	return nil
}

