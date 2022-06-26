//go:generate mockgen -destination=mocks/db.go -package mocks . DB
package repositories

import (
	"context"

	"github.com/addme96/simple-go-service/simple-service/database"
	"github.com/addme96/simple-go-service/simple-service/entities"
	"github.com/jackc/pgx/v4"
)

type DB interface {
	GetConn(ctx context.Context) (database.PgxConn, error)
}

type Resource struct {
	db DB
}

func NewResource(db DB) *Resource {
	return &Resource{db: db}
}

func (r Resource) Create(ctx context.Context, newResource entities.Resource) (int, error) {
	conn, err := r.db.GetConn(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close(ctx)
	stDesc, err := conn.Prepare(ctx, "createResource", "INSERT into resources (name) VALUES ($1) RETURNING id")
	if err != nil {
		return 0, err
	}
	row := conn.QueryRow(ctx, stDesc.Name, newResource.Name)
	var id int
	if err = row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r Resource) Read(ctx context.Context, id int) (*entities.Resource, error) {
	conn, err := r.db.GetConn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)
	stDesc, err := conn.Prepare(ctx, "readResource", "SELECT id, name FROM resources WHERE id=$1")
	if err != nil {
		return nil, err
	}
	var resource entities.Resource
	err = conn.QueryRow(ctx, stDesc.Name, id).Scan(&resource.ID, &resource.Name)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (r Resource) ReadAll(ctx context.Context) ([]entities.Resource, error) {
	conn, err := r.db.GetConn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)
	rows, err := conn.Query(ctx, "SELECT id, name FROM resources")
	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}
	resources := make([]entities.Resource, 0)
	if err == pgx.ErrNoRows {
		return resources, nil
	}
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
	conn, err := r.db.GetConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	stDesc, err := conn.Prepare(ctx, "updateResource", "UPDATE resources SET name = $1 WHERE id=$2")
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, stDesc.Name, newResource.Name, id)
	if err != nil {
		return err
	}
	return nil
}

func (r Resource) Delete(ctx context.Context, id int) error {
	conn, err := r.db.GetConn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	stDesc, err := conn.Prepare(ctx, "deleteResource", "DELETE FROM resources WHERE id=$1")
	if err != nil {
		return err
	}
	_, err = conn.Exec(ctx, stDesc.Name, id)
	if err != nil {
		return err
	}
	return nil
}
