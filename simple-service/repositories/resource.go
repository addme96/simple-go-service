package repositories

import (
	"context"

	"github.com/addme96/simple-go-service/simple-service/database"
	"github.com/addme96/simple-go-service/simple-service/entities"
)

type Resource struct {
	DB *database.DB
}

func NewResource(DB *database.DB) *Resource {
	return &Resource{DB: DB}
}

func (r Resource) Create(ctx context.Context, newResource entities.Resource) error {
	conn := r.DB.GetConn()
	defer conn.Close(ctx)
	stDesc, err := conn.Prepare(ctx, "createResource", "INSERT into resources (name) VALUES ($1)")
	if err != nil {
		return err
	}
	_, err = conn.Query(ctx, stDesc.Name, newResource.Name)
	if err != nil {
		return err
	}
	return nil
}

func (r Resource) Read(ctx context.Context, id int) (*entities.Resource, error) {
	conn := r.DB.GetConn()
	defer conn.Close(ctx)
	stDesc, err := conn.Prepare(ctx, "readResource", "SELECT id, name FROM resources WHERE id=$1")
	if err != nil {
		return nil, err
	}
	var resource entities.Resource
	err = conn.QueryRow(ctx, stDesc.Name, id).Scan(&resource)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (r Resource) ReadAll(ctx context.Context) ([]entities.Resource, error) {
	conn := r.DB.GetConn()
	defer conn.Close(ctx)
	rows, err := conn.Query(ctx, "SELECT id, name FROM resources")
	if err != nil {
		return nil, err
	}
	var resources []entities.Resource
	for rows.Next() {
		var resource entities.Resource
		err = rows.Scan(&resource.ID, &resource.Name)
		if err != nil {
			return nil, err
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func (r Resource) Update(ctx context.Context, id int, newResource entities.Resource) error {
	conn := r.DB.GetConn()
	defer conn.Close(ctx)
	stDesc, err := conn.Prepare(ctx, "updateResource", "UPDATE resources SET name = $1 WHERE id=$2")
	if err != nil {
		return err
	}
	_, err = conn.Query(ctx, stDesc.Name, newResource.Name, id)
	if err != nil {
		return err
	}
	return nil
}

func (r Resource) Delete(ctx context.Context, id int) error {
	conn := r.DB.GetConn()
	defer conn.Close(ctx)
	stDesc, err := conn.Prepare(ctx, "deleteResource", "DELETE FROM resources WHERE id=$1")
	if err != nil {
		return err
	}
	_, err = conn.Query(ctx, stDesc.Name, id)
	if err != nil {
		return err
	}
	return nil
}
